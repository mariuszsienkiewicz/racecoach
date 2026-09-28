<?php

namespace App\Exception\Chat;

/**
 * fit-chat rejected the request (4xx) after Symfony already authorized the user.
 */
final class FitChatRejectedException extends FitChatException
{
    public function __construct(
        string $message,
        public readonly int $statusCode,
    ) {
        parent::__construct($message);
    }
}
