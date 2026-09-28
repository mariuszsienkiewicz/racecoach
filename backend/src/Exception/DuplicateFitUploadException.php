<?php

namespace App\Exception;

use RuntimeException;

final class DuplicateFitUploadException extends RuntimeException
{
    public function __construct(
        public readonly int $existingActivityId,
        string $message = 'This .fit file was already uploaded.',
    ) {
        parent::__construct($message);
    }
}
