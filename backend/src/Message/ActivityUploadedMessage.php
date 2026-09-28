<?php

namespace App\Message;

/**
 * Queue event after a .fit was accepted.
 * Workers should fetch bytes from object storage using bucket + objectKey.
 */
final readonly class ActivityUploadedMessage
{
    public function __construct(
        public int $activityId,
        public int $userId,
        public string $originalFilename,
        public string $storageBucket,
        public string $objectKey,
        public string $checksumSha256,
        public int $fileSizeBytes,
        public string $uploadedAt,
    ) {
    }
}
