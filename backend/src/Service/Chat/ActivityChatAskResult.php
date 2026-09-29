<?php

namespace App\Service\Chat;

use App\Entity\ActivityChatMessage;

final readonly class ActivityChatAskResult
{
    public function __construct(
        public string $reply,
        public ActivityChatMessage $athleteMessage,
        public ActivityChatMessage $coachMessage,
    ) {
    }
}
