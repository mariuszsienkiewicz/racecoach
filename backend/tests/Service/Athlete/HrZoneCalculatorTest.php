<?php

namespace App\Tests\Service\Athlete;

use App\Service\Athlete\HrZoneCalculator;
use PHPUnit\Framework\TestCase;

final class HrZoneCalculatorTest extends TestCase
{
    public function testPercentOfHrMaxProducesFiveContiguousZones(): void
    {
        $zones = (new HrZoneCalculator())->fromPercentOfHrMax(200);

        self::assertCount(5, $zones);
        self::assertSame(1, $zones[0]['zone']);
        self::assertSame(100, $zones[0]['minBpm']);
        self::assertSame(120, $zones[0]['maxBpm']);
        self::assertSame(120, $zones[1]['minBpm']);
        self::assertSame(200, $zones[4]['maxBpm']);
    }

    public function testKarvonenUsesReserve(): void
    {
        $zones = (new HrZoneCalculator())->fromHeartRateReserve(190, 50);

        // Z2 upper at 70% HRR: 50 + 0.7*(190-50) = 148
        self::assertSame(148, $zones[1]['maxBpm']);
        self::assertSame(190, $zones[4]['maxBpm']);
    }

    public function testEstimateFromBirthYear(): void
    {
        $hrMax = (new HrZoneCalculator())->estimateHrMaxFromBirthYear(1990, 2026);
        self::assertSame(184, $hrMax);
    }
}
