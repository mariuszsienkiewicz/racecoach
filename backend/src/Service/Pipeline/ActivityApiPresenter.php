<?php

namespace App\Service\Pipeline;

use App\Entity\Activity;

final class ActivityApiPresenter
{
    public function __construct(
        private readonly ActivityReprocessEligibility $eligibility,
    ) {
    }

    /**
     * @param array<int, true>|null $recentIdSet
     *
     * @return array<string, mixed>
     */
    public function present(Activity $activity, ?array $recentIdSet = null): array
    {
        return [
            ...$activity->toApiArray(),
            ...$this->eligibility->describe($activity, $recentIdSet),
        ];
    }

    /**
     * @param list<Activity> $activities newest-first
     *
     * @return list<array<string, mixed>>
     */
    public function presentMany(array $activities): array
    {
        $window = $this->eligibility->recentIdSetFromList($activities);

        return array_map(
            fn (Activity $activity) => $this->present($activity, $window),
            $activities,
        );
    }
}
