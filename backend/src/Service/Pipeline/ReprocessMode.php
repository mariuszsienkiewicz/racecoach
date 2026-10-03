<?php

namespace App\Service\Pipeline;

enum ReprocessMode: string
{
    case FULL = 'full';
    case FROM_STRUCTURE = 'from_structure';
}
