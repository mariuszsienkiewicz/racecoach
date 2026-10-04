<?php

namespace App\Controller;

use App\Entity\Activity;
use App\Entity\ActivityChatMessage;
use App\Entity\User;
use App\Exception\ActivityCannotReprocessException;
use App\Exception\Chat\FitChatProtocolException;
use App\Exception\Chat\FitChatRejectedException;
use App\Exception\Chat\FitChatUnavailableException;
use App\Exception\Chat\InvalidChatMessageException;
use App\Exception\DuplicateFitUploadException;
use App\Repository\ActivityRepository;
use App\Service\Chat\ActivityChatService;
use App\Service\Fit\FitUploadService;
use App\Service\Pipeline\ActivityApiPresenter;
use App\Service\Pipeline\ActivityReprocessService;
use App\Service\Pipeline\ReprocessMode;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\ExpressionLanguage\Expression;
use Symfony\Component\HttpFoundation\File\UploadedFile;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\HttpFoundation\StreamedResponse;
use Symfony\Component\HttpKernel\Exception\BadRequestHttpException;
use Symfony\Component\Routing\Attribute\Route;
use Symfony\Component\Security\Http\Attribute\CurrentUser;
use Symfony\Component\Security\Http\Attribute\IsGranted;

final class ActivityController extends AbstractController
{
    private const MAX_MESSAGE_LENGTH = 2000;

    public function __construct(
        private readonly ActivityRepository $activityRepository,
        private readonly ActivityApiPresenter $activityApiPresenter,
        private readonly FitUploadService $fitUploadService,
        private readonly ActivityReprocessService $activityReprocessService,
        private readonly ActivityChatService $activityChatService,
    ) {
    }

    #[IsGranted('ROLE_USER')]
    #[Route('/api/activities', name: 'api_activities_list', methods: ['GET'])]
    public function list(
        #[CurrentUser] ?User $user,
    ): JsonResponse {
        if (!$user instanceof User) {
            return $this->json(['message' => 'Unauthorized'], Response::HTTP_UNAUTHORIZED);
        }

        $activities = $this->activityRepository->findRecentForUser($user, 100);

        return $this->json([
            'activities' => $this->activityApiPresenter->presentMany($activities),
        ]);
    }

    #[IsGranted(new Expression("is_granted('ROLE_USER') or is_granted('ROLE_WORKER')"))]
    #[Route(
        '/api/activities/{id}',
        name: 'api_activity_get',
        methods: ['GET'],
        requirements: ['id' => '\d+'],
    )]
    public function getActivity(
        int $id,
    ): JsonResponse {
        $activity = $this->activityRepository->find($id);
        if (!$activity instanceof Activity) {
            return $this->json(['message' => 'Activity not found'], Response::HTTP_NOT_FOUND);
        }

        if ($this->isGranted('ROLE_WORKER')) {
            return $this->json($activity->toApiArray());
        }

        $user = $this->getUser();
        if (!$user instanceof User || $activity->getUser()->getId() !== $user->getId()) {
            return $this->json(['message' => 'Activity not found'], Response::HTTP_NOT_FOUND);
        }

        return $this->json($this->activityApiPresenter->present($activity));
    }

    #[IsGranted('ROLE_USER')]
    #[Route('/api/activities/fit', name: 'api_activities_fit_upload', methods: ['POST'])]
    public function upload(Request $request, #[CurrentUser] ?User $user): JsonResponse
    {
        if (!$user instanceof User) {
            return $this->json(['message' => 'Unauthorized'], Response::HTTP_UNAUTHORIZED);
        }

        $uploaded = $request->files->get('fitFile');
        if (!$uploaded instanceof UploadedFile) {
            throw new BadRequestHttpException('Missing multipart field "fitFile" with a .fit file.');
        }

        try {
            $activity = $this->fitUploadService->upload($user, $uploaded);
        } catch (DuplicateFitUploadException $exception) {
            return $this->json([
                'message' => $exception->getMessage(),
                'existingActivityId' => (string) $exception->existingActivityId,
            ], Response::HTTP_CONFLICT);
        }

        return $this->json($activity->toApiArray(), Response::HTTP_CREATED);
    }

    #[IsGranted('ROLE_USER')]
    #[Route(
        '/api/activities/{id}/reprocess',
        name: 'api_activities_reprocess',
        methods: ['POST'],
        requirements: ['id' => '\d+'],
    )]
    public function reprocess(
        int $id,
        Request $request,
        #[CurrentUser] ?User $user,
    ): JsonResponse {
        if (!$user instanceof User) {
            return $this->json(['message' => 'Unauthorized'], Response::HTTP_UNAUTHORIZED);
        }

        $activity = $this->findOwnedActivity($id, $user);
        if (null === $activity) {
            return $this->json(['message' => 'Activity not found'], Response::HTTP_NOT_FOUND);
        }

        try {
            $mode = $this->parseReprocessMode($request);
        } catch (\InvalidArgumentException $exception) {
            return $this->json(['message' => $exception->getMessage()], Response::HTTP_BAD_REQUEST);
        }

        try {
            $activity = $this->activityReprocessService->reprocess($activity, $mode);
        } catch (ActivityCannotReprocessException $exception) {
            return $this->json(['message' => $exception->getMessage()], Response::HTTP_CONFLICT);
        }

        return $this->json($this->activityApiPresenter->present($activity), Response::HTTP_ACCEPTED);
    }

    #[IsGranted('ROLE_USER')]
    #[Route(
        '/api/activities/{id}/chat',
        name: 'api_activities_chat_get',
        methods: ['GET'],
        requirements: ['id' => '\d+'],
    )]
    public function listChat(
        int $id,
        #[CurrentUser] ?User $user,
    ): JsonResponse {
        if (!$user instanceof User) {
            return $this->json(['message' => 'Unauthorized'], Response::HTTP_UNAUTHORIZED);
        }

        $activity = $this->findOwnedActivity($id, $user);
        if (null === $activity) {
            return $this->json(['message' => 'Activity not found'], Response::HTTP_NOT_FOUND);
        }

        $messages = array_map(
            static fn (ActivityChatMessage $message) => $message->toApiArray(),
            $this->activityChatService->list($activity),
        );

        return $this->json([
            'activityId' => $activity->getId(),
            'messages' => $messages,
        ]);
    }

    #[IsGranted('ROLE_USER')]
    #[Route(
        '/api/activities/{id}/chat',
        name: 'api_activities_chat_post',
        methods: ['POST'],
        requirements: ['id' => '\d+'],
    )]
    public function chat(
        int $id,
        Request $request,
        #[CurrentUser] ?User $user,
    ): Response {
        if (!$user instanceof User) {
            return $this->json(['message' => 'Unauthorized'], Response::HTTP_UNAUTHORIZED);
        }

        $activity = $this->findOwnedActivity($id, $user);
        if (null === $activity) {
            return $this->json(['message' => 'Activity not found'], Response::HTTP_NOT_FOUND);
        }

        if (Activity::STATUS_READY !== $activity->getStatus()) {
            return $this->json(
                ['message' => 'Activity is not ready for chat yet'],
                Response::HTTP_CONFLICT,
            );
        }

        $featuresObjectKey = $activity->getFeaturesObjectKey();
        if (null === $featuresObjectKey || '' === $featuresObjectKey) {
            return $this->json(
                ['message' => 'Activity features are not available yet'],
                Response::HTTP_CONFLICT,
            );
        }

        try {
            $message = $this->parseChatMessage($request);
        } catch (InvalidChatMessageException $exception) {
            return $this->json(['message' => $exception->getMessage()], Response::HTTP_BAD_REQUEST);
        }

        return new StreamedResponse(
            function () use ($activity, $message): void {
                $this->emitChatStream($activity, $message);
            },
            Response::HTTP_OK,
            [
                'Content-Type' => 'text/event-stream',
                'Cache-Control' => 'no-cache',
                'Connection' => 'keep-alive',
                'X-Accel-Buffering' => 'no',
            ],
        );
    }

    private function emitChatStream(Activity $activity, string $message): void
    {
        while (ob_get_level() > 0) {
            ob_end_flush();
        }

        try {
            $result = $this->activityChatService->ask(
                $activity,
                $message,
                function (string $text): void {
                    $this->writeSse('token', ['text' => $text]);
                },
            );

            $this->writeSse('done', [
                'reply' => $result->reply,
                'activityId' => $activity->getId(),
                'messages' => [
                    $result->athleteMessage->toApiArray(),
                    $result->coachMessage->toApiArray(),
                ],
            ]);
        } catch (FitChatUnavailableException) {
            $this->writeSse('error', ['message' => 'AI coach unavailable']);
        } catch (FitChatRejectedException $exception) {
            $this->writeSse('error', ['message' => $this->publicFitChatRejectedMessage($exception)]);
        } catch (FitChatProtocolException) {
            $this->writeSse('error', ['message' => 'Invalid chat response']);
        } catch (\Throwable) {
            $this->writeSse('error', ['message' => 'AI coach unavailable']);
        }
    }

    private function publicFitChatRejectedMessage(FitChatRejectedException $exception): string
    {
        $detail = $exception->getMessage();
        if (str_contains(strtolower($detail), 'message is too long')) {
            return 'Your message is too long. Please shorten it and try again.';
        }

        return 'Could not process your message. Please try again.';
    }

    /**
     * @param array<string, mixed> $payload
     */
    private function writeSse(string $event, array $payload): void
    {
        echo sprintf("event: %s\ndata: %s\n\n", $event, json_encode($payload, JSON_THROW_ON_ERROR));
        if (function_exists('flush')) {
            flush();
        }
    }

    private function findOwnedActivity(int $id, User $user): ?Activity
    {
        $activity = $this->activityRepository->find($id);
        if (!$activity instanceof Activity || $activity->getUser()->getId() !== $user->getId()) {
            return null;
        }

        return $activity;
    }

    private function parseReprocessMode(Request $request): ReprocessMode
    {
        $raw = $request->getContent();
        if ('' === trim($raw)) {
            return ReprocessMode::FULL;
        }

        try {
            $payload = json_decode($raw, true, 512, \JSON_THROW_ON_ERROR);
        } catch (\JsonException) {
            throw new \InvalidArgumentException('Invalid JSON');
        }

        if (!\is_array($payload)) {
            throw new \InvalidArgumentException('Invalid JSON');
        }

        if (!\array_key_exists('mode', $payload) || null === $payload['mode'] || '' === $payload['mode']) {
            return ReprocessMode::FULL;
        }

        if (!\is_string($payload['mode'])) {
            throw new \InvalidArgumentException('mode must be a string');
        }

        $mode = ReprocessMode::tryFrom($payload['mode']);
        if (!$mode instanceof ReprocessMode) {
            throw new \InvalidArgumentException('mode must be full or from_structure');
        }

        return $mode;
    }

    /**
     * @throws InvalidChatMessageException when the JSON body is missing or invalid
     */
    private function parseChatMessage(Request $request): string
    {
        $payload = json_decode($request->getContent(), true);
        if (!\is_array($payload)) {
            throw new InvalidChatMessageException('Invalid JSON');
        }

        $raw = $payload['message'] ?? null;
        if (!\is_string($raw)) {
            throw new InvalidChatMessageException('message is required');
        }

        $message = trim($raw);
        if ('' === $message) {
            throw new InvalidChatMessageException('message is required');
        }

        if (mb_strlen($message) > self::MAX_MESSAGE_LENGTH) {
            throw new InvalidChatMessageException('Your message is too long. Please shorten it and try again.');
        }

        return $message;
    }
}
