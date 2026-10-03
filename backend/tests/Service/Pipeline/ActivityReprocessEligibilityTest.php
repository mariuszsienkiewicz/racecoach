<?php

namespace App\Tests\Service\Pipeline;

use App\Entity\Activity;
use App\Entity\User;
use App\Exception\ActivityCannotReprocessException;
use App\Repository\ActivityRepository;
use App\Service\Pipeline\ActivityReprocessEligibility;
use App\Service\Pipeline\AnalysisVersions;
use PHPUnit\Framework\TestCase;

final class ActivityReprocessEligibilityTest extends TestCase
{
    public function testOutdatedReadyActivityInWindowIsAvailable(): void
    {
        $activity = $this->activity(status: Activity::STATUS_READY, version: null, id: 1);
        $eligibility = new ActivityReprocessEligibility($this->createStub(ActivityRepository::class));

        $desc = $eligibility->describe($activity, [1 => true]);

        self::assertTrue($desc['reprocessAvailable']);
        self::assertSame('outdated_analysis', $desc['reprocessReason']);
    }

    public function testCurrentVersionIsNotAvailable(): void
    {
        $activity = $this->activity(status: Activity::STATUS_READY, version: AnalysisVersions::CURRENT, id: 1);
        $eligibility = new ActivityReprocessEligibility($this->createStub(ActivityRepository::class));

        $desc = $eligibility->describe($activity, [1 => true]);

        self::assertFalse($desc['reprocessAvailable']);
        self::assertSame('already_current', $desc['reprocessReason']);
    }

    public function testOutsideWindowIsNotAvailable(): void
    {
        $activity = $this->activity(status: Activity::STATUS_READY, version: 1, id: 99);
        $eligibility = new ActivityReprocessEligibility($this->createStub(ActivityRepository::class));

        $desc = $eligibility->describe($activity, [1 => true]);

        self::assertFalse($desc['reprocessAvailable']);
        self::assertSame('outside_window', $desc['reprocessReason']);
    }

    public function testStuckAnalyzingIsAvailable(): void
    {
        $activity = $this->activity(status: Activity::STATUS_ANALYZING, version: null, id: 1);
        $eligibility = new ActivityReprocessEligibility($this->createStub(ActivityRepository::class));

        $desc = $eligibility->describe($activity, [1 => true]);

        self::assertTrue($desc['reprocessAvailable']);
        self::assertSame('stale_pipeline', $desc['reprocessReason']);
    }

    public function testAssertEligibleThrowsForCurrent(): void
    {
        $activity = $this->activity(status: Activity::STATUS_READY, version: AnalysisVersions::CURRENT, id: 1);
        $eligibility = new ActivityReprocessEligibility($this->createStub(ActivityRepository::class));

        $this->expectException(ActivityCannotReprocessException::class);
        $eligibility->assertEligible($activity, [1 => true]);
    }

    public function testRecentIdSetFromListHonorsLimit(): void
    {
        $activities = [];
        for ($i = 1; $i <= AnalysisVersions::REPROCESS_RECENT_LIMIT + 5; ++$i) {
            $activities[] = $this->activity(status: Activity::STATUS_READY, version: 1, id: $i);
        }

        $eligibility = new ActivityReprocessEligibility($this->createStub(ActivityRepository::class));
        $set = $eligibility->recentIdSetFromList($activities);

        self::assertCount(AnalysisVersions::REPROCESS_RECENT_LIMIT, $set);
        self::assertArrayHasKey(1, $set);
        self::assertArrayHasKey(AnalysisVersions::REPROCESS_RECENT_LIMIT, $set);
        self::assertArrayNotHasKey(AnalysisVersions::REPROCESS_RECENT_LIMIT + 1, $set);
    }

    private function activity(string $status, ?int $version, int $id): Activity
    {
        $user = (new User())->setEmail('coach@example.com')->setPassword('x');
        $this->setEntityId($user, 7);

        $activity = (new Activity())
            ->setUser($user)
            ->setTitle('Race')
            ->setStatus($status)
            ->setStartedAt(new \DateTimeImmutable('2024-05-01T10:00:00+00:00'))
            ->setOriginalFilename('race.fit')
            ->setObjectKey('users/1/fits/race.fit')
            ->setStorageBucket('racecoach-fits')
            ->setChecksumSha256(str_repeat('a', 64))
            ->setAnalysisVersion($version);

        $this->setEntityId($activity, $id);

        return $activity;
    }

    private function setEntityId(object $entity, int $id): void
    {
        $ref = new \ReflectionProperty($entity, 'id');
        $ref->setValue($entity, $id);
    }
}
