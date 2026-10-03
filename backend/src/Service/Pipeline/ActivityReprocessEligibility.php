<?php

namespace App\Service\Pipeline;

use App\Entity\Activity;
use App\Entity\User;
use App\Exception\ActivityCannotReprocessException;
use App\Repository\ActivityRepository;

final class ActivityReprocessEligibility
{
    public function __construct(
        private readonly ActivityRepository $activityRepository,
    ) {
    }

    /**
     * @param array<int, true>|null $recentIdSet precomputed window for list endpoints
     *
     * @return array{reprocessAvailable: bool, reprocessReason: string}
     */
    public function describe(Activity $activity, ?array $recentIdSet = null): array
    {
        $reason = $this->reason($activity, $recentIdSet);

        return [
            'reprocessAvailable' => 'outdated_analysis' === $reason,
            'reprocessReason' => $reason,
        ];
    }

    /**
     * @param list<Activity> $activities newest-first
     *
     * @return array<int, true>
     */
    public function recentIdSetFromList(array $activities): array
    {
        $set = [];
        foreach (\array_slice($activities, 0, AnalysisVersions::REPROCESS_RECENT_LIMIT) as $activity) {
            $id = $activity->getId();
            if (null !== $id) {
                $set[$id] = true;
            }
        }

        return $set;
    }

    /**
     * @return array<int, true>
     */
    public function recentIdSetForUser(User $user): array
    {
        return $this->recentIdSetFromList(
            $this->activityRepository->findRecentForUser($user, AnalysisVersions::REPROCESS_RECENT_LIMIT),
        );
    }

    /**
     * @param array<int, true>|null $recentIdSet
     */
    public function assertEligible(Activity $activity, ?array $recentIdSet = null): void
    {
        $reason = $this->reason($activity, $recentIdSet);
        if ('outdated_analysis' === $reason) {
            return;
        }

        throw new ActivityCannotReprocessException(match ($reason) {
            'already_current' => 'Analysis is already up to date.', 'outside_window' => 'Only recent activities can be reprocessed.', 'not_ready' => 'Activity is not ready for reprocess.', 'missing_fit' => 'Activity has no FIT object to reprocess.', default => 'Activity cannot be reprocessed.',
        });
    }

    /**
     * @param array<int, true>|null $recentIdSet
     */
    private function reason(Activity $activity, ?array $recentIdSet): string
    {
        if (null === $activity->getObjectKey() || '' === $activity->getObjectKey()) {
            return 'missing_fit';
        }
        if (Activity::STATUS_READY !== $activity->getStatus()) {
            return 'not_ready';
        }

        $version = $activity->getAnalysisVersion() ?? 0;
        if ($version >= AnalysisVersions::CURRENT) {
            return 'already_current';
        }

        $id = $activity->getId();
        if (null === $id) {
            return 'not_ready';
        }

        $window = $recentIdSet ?? $this->recentIdSetForUser($activity->getUser());
        if (!isset($window[$id])) {
            return 'outside_window';
        }

        return 'outdated_analysis';
    }
}
