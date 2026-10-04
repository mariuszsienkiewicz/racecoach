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
            'reprocessAvailable' => \in_array($reason, ['outdated_analysis', 'stale_pipeline'], true),
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
        if (\in_array($reason, ['outdated_analysis', 'stale_pipeline'], true)) {
            return;
        }

        throw new ActivityCannotReprocessException(match ($reason) {
            'already_current' => 'Analysis is already up to date.', 'outside_window' => 'Only recent activities can be reprocessed.', 'pipeline_in_progress' => 'Analysis is still running. Try again if it stays stuck.', 'not_ready' => 'Activity is not ready for reprocess.', 'missing_fit' => 'Activity has no FIT object to reprocess.', default => 'Activity cannot be reprocessed.',
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

        $id = $activity->getId();
        if (null === $id) {
            return 'not_ready';
        }

        $window = $recentIdSet ?? $this->recentIdSetForUser($activity->getUser());
        if (!isset($window[$id])) {
            return 'outside_window';
        }

        $status = $activity->getStatus();
        if (Activity::STATUS_FAILED === $status) {
            return 'stale_pipeline';
        }
        if (\in_array($status, [Activity::STATUS_ANALYZING, Activity::STATUS_UPLOADED], true)) {
            return $this->isPipelineHeartbeatStale($activity) ? 'stale_pipeline' : 'pipeline_in_progress';
        }
        if (Activity::STATUS_READY !== $status) {
            return 'not_ready';
        }

        $version = $activity->getAnalysisVersion() ?? 0;
        if ($version >= AnalysisVersions::CURRENT) {
            return 'already_current';
        }

        return 'outdated_analysis';
    }

    private function isPipelineHeartbeatStale(Activity $activity): bool
    {
        $heartbeat = $activity->getPipelineHeartbeatAt();
        if (null === $heartbeat) {
            return true;
        }

        $threshold = new \DateTimeImmutable('-'.AnalysisVersions::STALE_PIPELINE_AFTER);

        return $heartbeat <= $threshold;
    }
}
