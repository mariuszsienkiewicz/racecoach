<?php

namespace App\Exception\Chat;

use RuntimeException;

/**
 * Base failure talking to the internal fit-chat service.
 */
abstract class FitChatException extends RuntimeException
{
}
