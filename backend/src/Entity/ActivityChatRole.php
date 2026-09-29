<?php

namespace App\Entity;

enum ActivityChatRole: string
{
    case ATHLETE = 'athlete';
    case COACH = 'coach';
}
