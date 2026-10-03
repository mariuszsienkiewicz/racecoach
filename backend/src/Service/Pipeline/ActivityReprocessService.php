<?php

namespace App\Service\Pipeline;

use App\Entity\Activity;
use App\Exception\ActivityCannotReprocessException;
use App\Message\ActivityMetricsReadyMessage;
use App\Message\ActivityUploadedMessage;
use Doctrine\ORM\EntityManagerInterface;
use Symfony\Component\Messenger\Bridge\Amqp\Transport\AmqpStamp;
use Symfony\Component\Messenger\MessageBusInterface;

final class ActivityReprocessService
{
    public function __construct(
        private readonly EntityManagerInterface $entityManager,
        private readonly MessageBusInterface $messageBus,
    ) {
    }

    public function reprocess(Activity $activity, ReprocessMode $mode = ReprocessMode::FULL): Activity
    {
        $this->assertReprocessable($activity, $mode);

        $activity
            ->setStructureObjectKey(null)
            ->setFeaturesObjectKey(null)
            ->setSummaryObjectKey(null)
            ->setSummary('Queued for analysis.');

        if (ReprocessMode::FULL === $mode) {
            $activity
                ->setStatus(Activity::STATUS_UPLOADED)
                ->setMetricsObjectKey(null);
        } else {
            $activity->setStatus(Activity::STATUS_ANALYZING);
        }

        $this->entityManager->flush();

        match ($mode) {
            ReprocessMode::FULL => $this->publishUploaded($activity),
            ReprocessMode::FROM_STRUCTURE => $this->publishMetricsReady($activity),
        };

        return $activity;
    }

    private function assertReprocessable(Activity $activity, ReprocessMode $mode): void
    {
        if (null === $activity->getId()) {
            throw new ActivityCannotReprocessException('Activity must be persisted.');
        }
        if (null === $activity->getObjectKey() || '' === $activity->getObjectKey()) {
            throw new ActivityCannotReprocessException('Activity has no FIT object to reprocess.');
        }
        if (null === $activity->getStorageBucket() || '' === $activity->getStorageBucket()) {
            throw new ActivityCannotReprocessException('Activity has no storage bucket.');
        }
        if (null === $activity->getChecksumSha256() || '' === $activity->getChecksumSha256()) {
            throw new ActivityCannotReprocessException('Activity has no checksum.');
        }
        if (null === $activity->getUser()->getId()) {
            throw new ActivityCannotReprocessException('Activity owner must be persisted.');
        }

        if (ReprocessMode::FROM_STRUCTURE === $mode) {
            if (null === $activity->getMetricsObjectKey() || '' === $activity->getMetricsObjectKey()) {
                throw new ActivityCannotReprocessException('from_structure requires existing metrics; use mode=full instead.');
            }
        }
    }

    private function publishUploaded(Activity $activity): void
    {
        $this->messageBus->dispatch(
            new ActivityUploadedMessage(
                activityId: (int) $activity->getId(),
                userId: (int) $activity->getUser()->getId(),
                originalFilename: $activity->getOriginalFilename(),
                storageBucket: (string) $activity->getStorageBucket(),
                objectKey: (string) $activity->getObjectKey(),
                checksumSha256: (string) $activity->getChecksumSha256(),
                fileSizeBytes: $activity->getFileSizeBytes(),
                uploadedAt: $activity->getCreatedAt()->format(\DateTimeInterface::ATOM),
            ),
            [new AmqpStamp('activity.uploaded')],
        );
    }

    private function publishMetricsReady(Activity $activity): void
    {
        $this->messageBus->dispatch(
            new ActivityMetricsReadyMessage(
                activityId: (int) $activity->getId(),
                userId: (int) $activity->getUser()->getId(),
                storageBucket: (string) $activity->getStorageBucket(),
                objectKey: (string) $activity->getObjectKey(),
                metricsObjectKey: (string) $activity->getMetricsObjectKey(),
                distanceM: $activity->getDistanceM(),
                durationSec: $activity->getDurationSec(),
                avgHeartRate: $activity->getAvgHeartRate(),
                readyAt: (new \DateTimeImmutable('now', new \DateTimeZone('UTC')))->format(\DateTimeInterface::ATOM),
            ),
            [new AmqpStamp('activity.metrics.ready')],
        );
    }
}
