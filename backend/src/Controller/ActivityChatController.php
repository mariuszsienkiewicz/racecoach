<?php

namespace App\Controller;

use App\Entity\Activity;
use App\Entity\ActivityChatMessage;
use App\Entity\User;
use App\Exception\Chat\FitChatProtocolException;
use App\Exception\Chat\FitChatRejectedException;
use App\Exception\Chat\FitChatUnavailableException;
use App\Exception\Chat\InvalidChatMessageException;
use App\Repository\ActivityRepository;
use App\Service\Chat\ActivityChatService;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\HttpFoundation\StreamedResponse;
use Symfony\Component\Routing\Attribute\Route;
use Symfony\Component\Security\Http\Attribute\CurrentUser;
use Symfony\Component\Security\Http\Attribute\IsGranted;

final class ActivityChatController extends AbstractController
{
    private const MAX_MESSAGE_LENGTH = 2000;

    public function __construct(
        private readonly ActivityChatService $activityChatService,
        private readonly ActivityRepository $activityRepository,
    ) {
    }

    #[IsGranted('ROLE_USER')]
    #[Route(
        '/api/activities/{id}/chat',
        name: 'api_activities_chat_get',
        methods: ['GET'],
        requirements: ['id' => '\d+'],
    )]
    public function list(
        int $id,
        #[CurrentUser] ?User $user,
    ): JsonResponse {
        if (!$user instanceof User) {
            return $this->json(['message' => 'Unauthorized'], Response::HTTP_UNAUTHORIZED);
        }

        $activity = $this->findOwnedActivity($id, $user);
        if ($activity === null) {
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
        if ($activity === null) {
            return $this->json(['message' => 'Activity not found'], Response::HTTP_NOT_FOUND);
        }

        if ($activity->getStatus() !== Activity::STATUS_READY) {
            return $this->json(
                ['message' => 'Activity is not ready for chat yet'],
                Response::HTTP_CONFLICT,
            );
        }

        $featuresObjectKey = $activity->getFeaturesObjectKey();
        if ($featuresObjectKey === null || $featuresObjectKey === '') {
            return $this->json(
                ['message' => 'Activity features are not available yet'],
                Response::HTTP_CONFLICT,
            );
        }

        try {
            $message = $this->parseMessage($request);
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
            $this->writeSse('error', ['message' => $exception->getMessage()]);
        } catch (FitChatProtocolException) {
            $this->writeSse('error', ['message' => 'Invalid chat response']);
        } catch (\Throwable) {
            $this->writeSse('error', ['message' => 'AI coach unavailable']);
        }
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
        if (!$activity instanceof Activity || $activity->getUser()?->getId() !== $user->getId()) {
            return null;
        }

        return $activity;
    }

    /**
     * @throws InvalidChatMessageException when the JSON body is missing or invalid
     */
    private function parseMessage(Request $request): string
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
        if ($message === '') {
            throw new InvalidChatMessageException('message is required');
        }

        if (mb_strlen($message) > self::MAX_MESSAGE_LENGTH) {
            throw new InvalidChatMessageException(
                sprintf('message is too long (max %d characters)', self::MAX_MESSAGE_LENGTH),
            );
        }

        return $message;
    }
}
