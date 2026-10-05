<?php

namespace App\Service\Athlete;

/**
 * Materializes 5 HR zones as bpm bounds.
 *
 * Default percent bands (coach-standard):
 * Z1 50–60, Z2 60–70, Z3 70–80, Z4 80–90, Z5 90–100.
 */
final class HrZoneCalculator
{
    /** @var list<array{0: float, 1: float}> */
    private const PCT_BANDS = [
        [0.50, 0.60],
        [0.60, 0.70],
        [0.70, 0.80],
        [0.80, 0.90],
        [0.90, 1.00],
    ];

    /**
     * @return list<array{zone: int, minBpm: int, maxBpm: int}>
     */
    public function fromPercentOfHrMax(int $hrMax): array
    {
        return $this->fromPercentBands($hrMax, null);
    }

    /**
     * Karvonen / %HRR: target = hrRest + pct * (hrMax - hrRest).
     *
     * @return list<array{zone: int, minBpm: int, maxBpm: int}>
     */
    public function fromHeartRateReserve(int $hrMax, int $hrRest): array
    {
        return $this->fromPercentBands($hrMax, $hrRest);
    }

    public function estimateHrMaxFromBirthYear(int $birthYear, ?int $currentYear = null): int
    {
        $year = $currentYear ?? (int) (new \DateTimeImmutable('now'))->format('Y');
        $age = max(10, min(100, $year - $birthYear));

        return max(120, min(230, 220 - $age));
    }

    /**
     * @return list<array{zone: int, minBpm: int, maxBpm: int}>
     */
    private function fromPercentBands(int $hrMax, ?int $hrRest): array
    {
        $zones = [];
        foreach (self::PCT_BANDS as $index => [$loPct, $hiPct]) {
            $zone = $index + 1;
            $min = $this->bpmAt($loPct, $hrMax, $hrRest);
            $max = $this->bpmAt($hiPct, $hrMax, $hrRest);
            if (5 === $zone) {
                $max = $hrMax;
            }
            if ($max < $min) {
                $max = $min;
            }
            $zones[] = [
                'zone' => $zone,
                'minBpm' => $min,
                'maxBpm' => $max,
            ];
        }

        for ($i = 1; $i < 5; ++$i) {
            $zones[$i]['minBpm'] = $zones[$i - 1]['maxBpm'];
        }

        return $zones;
    }

    private function bpmAt(float $pct, int $hrMax, ?int $hrRest): int
    {
        if (null !== $hrRest) {
            return (int) round($hrRest + $pct * ($hrMax - $hrRest));
        }

        return (int) round($pct * $hrMax);
    }
}
