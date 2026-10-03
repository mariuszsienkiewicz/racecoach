<?php

namespace App\Message;

/**
 * Queue event to (re)start the pipeline after metrics.
 */
final readonly class ActivityMetricsReadyMessage
{
    public function __construct(
        public int $activityId,
        public int $userId,
        public string $storageBucket,
        public string $objectKey,
        public string $metricsObjectKey,
        public int $distanceM,
        public int $durationSec,
        public ?int $avgHeartRate,
        public string $readyAt,
    ) {
    }
}
