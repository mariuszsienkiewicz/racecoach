<?php

namespace App\Service\Chat;

use App\Entity\Activity;
use App\Entity\ActivityChatMessage;
use App\Entity\ActivityChatRole;
use App\Repository\ActivityChatMessageRepository;
use App\Service\Storage\ObjectStorage;
use Doctrine\ORM\EntityManagerInterface;

final class ActivityChatService
{
    private const HISTORY_LIMIT = 12;

    public function __construct(
        private readonly ActivityChatMessageRepository $activityChatMessageRepository,
        private readonly EntityManagerInterface $entityManager,
        private readonly ObjectStorage $objectStorage,
        private readonly FitChatClient $fitChatClient,
    ) {
    }

    /**
     * @return list<ActivityChatMessage>
     */
    public function list(Activity $activity): array
    {
        return $this->activityChatMessageRepository->findByActivity($activity);
    }

    /**
     * Persist athlete, stream coach tokens via $onToken, persist coach on success.
     *
     * @param callable(string):void $onToken
     */
    public function ask(Activity $activity, string $message, callable $onToken): ActivityChatAskResult
    {
        $user = $activity->getUser();
        $activityId = $activity->getId();
        $featuresObjectKey = $activity->getFeaturesObjectKey();
        if (null === $activityId) {
            throw new \InvalidArgumentException('Activity must be persisted and owned.');
        }
        if (null === $featuresObjectKey || '' === $featuresObjectKey) {
            throw new \InvalidArgumentException('Activity featuresObjectKey is required for chat.');
        }

        $history = array_map(
            static fn (ActivityChatMessage $chatMessage): array => [
                'role' => $chatMessage->getRole()->value,
                'content' => $chatMessage->getContent(),
            ],
            $this->activityChatMessageRepository->findRecentByActivity($activity, self::HISTORY_LIMIT),
        );

        $athleteMessage = (new ActivityChatMessage())
            ->setActivity($activity)
            ->setRole(ActivityChatRole::ATHLETE)
            ->setContent($message);
        $this->entityManager->persist($athleteMessage);
        $this->entityManager->flush();

        $reply = $this->fitChatClient->askStream(new FitChatAskRequest(
            userId: (int) $user->getId(),
            activityId: $activityId,
            message: $message,
            storageBucket: $this->objectStorage->getBucket(),
            featuresObjectKey: $featuresObjectKey,
            summaryObjectKey: $activity->getSummaryObjectKey(),
            summary: $activity->getSummary(),
            history: $history,
        ), $onToken);

        $coachMessage = (new ActivityChatMessage())
            ->setActivity($activity)
            ->setRole(ActivityChatRole::COACH)
            ->setContent($reply->text);
        $this->entityManager->persist($coachMessage);
        $this->entityManager->flush();

        return new ActivityChatAskResult(
            reply: $reply->text,
            athleteMessage: $athleteMessage,
            coachMessage: $coachMessage,
        );
    }
}
