<?php

use Symfony\Component\Dotenv\Dotenv;

require dirname(__DIR__).'/vendor/autoload.php';

// AmqpStamp references ext-amqp constants, unit tests may run without it
if (!\defined('AMQP_NOPARAM')) {
    \define('AMQP_NOPARAM', 0);
}

(new Dotenv())->bootEnv(dirname(__DIR__).'/.env');

if ($_SERVER['APP_DEBUG']) {
    umask(0000);
}
