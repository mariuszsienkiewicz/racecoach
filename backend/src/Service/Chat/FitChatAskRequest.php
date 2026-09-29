<?php

namespace App\Service\Chat;

/**
 * Payload for the internal fit-chat inference service.
 */
final readonly class FitChatAskRequest
{
    public function __construct(
        public int $userId,
        public int $activityId,
        public string $message,
        public string $storageBucket,
        public string $featuresObjectKey,
        public ?string $summaryObjectKey = null,
        public ?string $summary = null,
        /** @var list<array{role: string, content: string}>|null */
        public ?array $history = null,
    ) {
    }
}
