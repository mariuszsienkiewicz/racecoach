<?php

namespace App\Tests\Service\Pipeline;

use App\Entity\Activity;
use App\Entity\User;
use App\Exception\ActivityCannotReprocessException;
use App\Message\ActivityMetricsReadyMessage;
use App\Message\ActivityUploadedMessage;
use App\Repository\ActivityRepository;
use App\Service\Pipeline\ActivityReprocessEligibility;
use App\Service\Pipeline\ActivityReprocessService;
use App\Service\Pipeline\AnalysisVersions;
use App\Service\Pipeline\ReprocessMode;
use Doctrine\ORM\EntityManagerInterface;
use PHPUnit\Framework\TestCase;
use Symfony\Component\Messenger\Envelope;
use Symfony\Component\Messenger\MessageBusInterface;

final class ActivityReprocessServiceTest extends TestCase
{
    public function testFullClearsArtifactsAndPublishesUploaded(): void
    {
        $activity = $this->readyActivity();
        $dispatched = [];

        $em = $this->createMock(EntityManagerInterface::class);
        $em->expects(self::once())->method('flush');

        $bus = $this->createMock(MessageBusInterface::class);
        $bus->expects(self::once())
            ->method('dispatch')
            ->willReturnCallback(function (object $message, array $stamps = []) use (&$dispatched): Envelope {
                $dispatched[] = [$message, $stamps];

                return new Envelope($message, $stamps);
            });

        $service = new ActivityReprocessService($em, $bus, $this->eligibilityAllowing($activity));
        $result = $service->reprocess($activity, ReprocessMode::FULL);

        self::assertSame(Activity::STATUS_UPLOADED, $result->getStatus());
        self::assertNull($result->getMetricsObjectKey());
        self::assertNull($result->getStructureObjectKey());
        self::assertNull($result->getFeaturesObjectKey());
        self::assertNull($result->getSummaryObjectKey());
        self::assertNull($result->getAnalysisVersion());
        self::assertSame('Queued for analysis.', $result->getSummary());

        self::assertCount(1, $dispatched);
        self::assertInstanceOf(ActivityUploadedMessage::class, $dispatched[0][0]);
        self::assertSame(42, $dispatched[0][0]->activityId);
    }

    public function testFromStructureKeepsMetricsAndPublishesMetricsReady(): void
    {
        $activity = $this->readyActivity();
        $dispatched = [];

        $em = $this->createMock(EntityManagerInterface::class);
        $em->expects(self::once())->method('flush');

        $bus = $this->createMock(MessageBusInterface::class);
        $bus->expects(self::once())
            ->method('dispatch')
            ->willReturnCallback(function (object $message, array $stamps = []) use (&$dispatched): Envelope {
                $dispatched[] = $message;

                return new Envelope($message, $stamps);
            });

        $service = new ActivityReprocessService($em, $bus, $this->eligibilityAllowing($activity));
        $result = $service->reprocess($activity, ReprocessMode::FROM_STRUCTURE);

        self::assertSame(Activity::STATUS_ANALYZING, $result->getStatus());
        self::assertSame('users/1/fits/race.metrics.json', $result->getMetricsObjectKey());
        self::assertNull($result->getAnalysisVersion());

        self::assertCount(1, $dispatched);
        self::assertInstanceOf(ActivityMetricsReadyMessage::class, $dispatched[0]);
        self::assertSame(5070, $dispatched[0]->distanceM);
    }

    public function testFromStructureWithoutMetricsFails(): void
    {
        $activity = $this->readyActivity();
        $activity->setMetricsObjectKey(null);

        $service = new ActivityReprocessService(
            $this->createStub(EntityManagerInterface::class),
            $this->createStub(MessageBusInterface::class),
            $this->eligibilityAllowing($activity),
        );

        $this->expectException(ActivityCannotReprocessException::class);
        $service->reprocess($activity, ReprocessMode::FROM_STRUCTURE);
    }

    public function testEligibilityFailureStopsReprocess(): void
    {
        $activity = $this->readyActivity();
        $activity->setAnalysisVersion(AnalysisVersions::CURRENT);

        $em = $this->createMock(EntityManagerInterface::class);
        $em->expects(self::never())->method('flush');

        $service = new ActivityReprocessService(
            $em,
            $this->createStub(MessageBusInterface::class),
            $this->eligibilityAllowing($activity),
        );

        $this->expectException(ActivityCannotReprocessException::class);
        $service->reprocess($activity, ReprocessMode::FULL);
    }

    private function eligibilityAllowing(Activity $activity): ActivityReprocessEligibility
    {
        $repo = $this->createStub(ActivityRepository::class);
        $repo->method('findRecentForUser')->willReturn([$activity]);

        return new ActivityReprocessEligibility($repo);
    }

    private function readyActivity(): Activity
    {
        $user = (new User())->setEmail('coach@example.com')->setPassword('x');
        $this->setEntityId($user, 7);

        $activity = (new Activity())
            ->setUser($user)
            ->setTitle('Race')
            ->setStatus(Activity::STATUS_READY)
            ->setStartedAt(new \DateTimeImmutable('2024-05-01T10:00:00+00:00'))
            ->setOriginalFilename('race.fit')
            ->setObjectKey('users/1/fits/race.fit')
            ->setStorageBucket('racecoach-fits')
            ->setChecksumSha256(str_repeat('a', 64))
            ->setFileSizeBytes(1200)
            ->setDistanceM(5070)
            ->setDurationSec(1458)
            ->setAvgHeartRate(170)
            ->setAnalysisVersion(1)
            ->setMetricsObjectKey('users/1/fits/race.metrics.json')
            ->setStructureObjectKey('users/1/fits/race.structure.json')
            ->setFeaturesObjectKey('users/1/fits/race.features.json')
            ->setSummaryObjectKey('users/1/fits/race.summary.json')
            ->setSummary('Old coach note.');

        $this->setEntityId($activity, 42);

        return $activity;
    }

    private function setEntityId(object $entity, int $id): void
    {
        $ref = new \ReflectionProperty($entity, 'id');
        $ref->setValue($entity, $id);
    }
}
