<?php

namespace App\Exception;

final class EmailAlreadyRegisteredException extends \RuntimeException
{
    public function __construct(string $message = 'Email is already registered.')
    {
        parent::__construct($message);
    }
}
