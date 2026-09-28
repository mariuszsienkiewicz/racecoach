<?php

namespace App\Controller;

use App\Entity\Activity;
use App\Entity\User;
use App\Exception\Chat\FitChatProtocolException;
use App\Exception\Chat\FitChatRejectedException;
use App\Exception\Chat\FitChatUnavailableException;
use App\Repository\ActivityRepository;
use App\Service\Chat\FitChatAskRequest;
use App\Service\Chat\FitChatClient;
use App\Service\Storage\ObjectStorage;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Attribute\Route;
use Symfony\Component\Security\Http\Attribute\CurrentUser;
use Symfony\Component\Security\Http\Attribute\IsGranted;

/**
 * Public chat gateway: JWT + ownership, then delegate inference to fit-chat.
 */
final class ActivityChatController extends AbstractController
{
    private const MAX_MESSAGE_LENGTH = 2000;

    public function __construct(
        private readonly FitChatClient $fitChatClient,
        private readonly ObjectStorage $objectStorage,
    ) {
    }

    #[IsGranted('ROLE_USER')]
    #[Route(
        '/api/activities/{id}/chat',
        name: 'api_activities_chat',
        methods: ['POST'],
        requirements: ['id' => '\d+'],
    )]
    public function chat(
        int $id,
        Request $request,
        #[CurrentUser] ?User $user,
        ActivityRepository $activityRepository,
    ): JsonResponse {
        if (!$user instanceof User) {
            return $this->json(['message' => 'Unauthorized'], Response::HTTP_UNAUTHORIZED);
        }

        $activity = $activityRepository->find($id);
        // Missing and foreign look the same - do not leak activity existence.
        if (!$activity instanceof Activity || $activity->getUser()?->getId() !== $user->getId()) {
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

        $message = $this->parseMessage($request);
        if ($message instanceof JsonResponse) {
            return $message;
        }

        try {
            $reply = $this->fitChatClient->ask(new FitChatAskRequest(
                userId: (int) $user->getId(),
                activityId: (int) $activity->getId(),
                message: $message,
                storageBucket: $this->objectStorage->getBucket(),
                featuresObjectKey: $featuresObjectKey,
                summaryObjectKey: $activity->getSummaryObjectKey(),
                summary: $activity->getSummary(),
            ));
        } catch (FitChatUnavailableException) {
            return $this->json(['message' => 'AI coach unavailable'], Response::HTTP_SERVICE_UNAVAILABLE);
        } catch (FitChatRejectedException $exception) {
            return $this->json(['message' => $exception->getMessage()], Response::HTTP_BAD_GATEWAY);
        } catch (FitChatProtocolException) {
            return $this->json(['message' => 'Invalid chat response'], Response::HTTP_BAD_GATEWAY);
        }

        return $this->json([
            'reply' => $reply->text,
            'activityId' => $activity->getId(),
        ]);
    }

    /**
     * @return string|JsonResponse validated message, or a 400 response
     */
    private function parseMessage(Request $request): string|JsonResponse
    {
        $payload = json_decode($request->getContent(), true);
        if (!\is_array($payload)) {
            return $this->json(['message' => 'Invalid JSON'], Response::HTTP_BAD_REQUEST);
        }

        $raw = $payload['message'] ?? null;
        if (!\is_string($raw)) {
            return $this->json(['message' => 'message is required'], Response::HTTP_BAD_REQUEST);
        }

        $message = trim($raw);
        if ($message === '') {
            return $this->json(['message' => 'message is required'], Response::HTTP_BAD_REQUEST);
        }

        if (mb_strlen($message) > self::MAX_MESSAGE_LENGTH) {
            return $this->json(
                ['message' => sprintf('message is too long (max %d characters)', self::MAX_MESSAGE_LENGTH)],
                Response::HTTP_BAD_REQUEST,
            );
        }

        return $message;
    }
}
