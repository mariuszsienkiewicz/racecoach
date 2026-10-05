<?php

namespace App\Service\Athlete;

use App\Entity\AthleteProfile;
use App\Entity\User;
use App\Exception\InvalidAthleteProfileException;
use App\Repository\AthleteProfileRepository;
use Doctrine\ORM\EntityManagerInterface;

final class AthleteProfileService
{
    public function __construct(
        private readonly EntityManagerInterface $entityManager,
        private readonly AthleteProfileRepository $athleteProfileRepository,
        private readonly HrZoneCalculator $hrZoneCalculator,
    ) {
    }

    public function getForUser(User $user): ?AthleteProfile
    {
        return $user->getAthleteProfile() ?? $this->athleteProfileRepository->findOneByUser($user);
    }

    /**
     * @param array<string, mixed> $payload
     */
    public function upsert(User $user, array $payload, bool $completeOnboarding = true): AthleteProfile
    {
        $profile = $this->getForUser($user);
        if (!$profile instanceof AthleteProfile) {
            $profile = new AthleteProfile();
            $profile->setUser($user);
            $user->setAthleteProfile($profile);
            $this->entityManager->persist($profile);
        }

        $birthYear = $this->optionalInt($payload, 'birthYear');
        if (null !== $birthYear) {
            $currentYear = (int) (new \DateTimeImmutable('now'))->format('Y');
            if ($birthYear < $currentYear - 100 || $birthYear > $currentYear - 10) {
                throw new InvalidAthleteProfileException('birthYear is out of range');
            }
            $profile->setBirthYear($birthYear);
        }

        if (\array_key_exists('sex', $payload)) {
            $sex = $payload['sex'];
            if (null === $sex || '' === $sex) {
                $profile->setSex(null);
            } elseif (!\is_string($sex) || !\in_array($sex, AthleteProfile::SEXES, true)) {
                throw new InvalidAthleteProfileException('sex is invalid');
            } else {
                $profile->setSex($sex);
            }
        }

        if (\array_key_exists('primaryGoal', $payload)) {
            $goal = $payload['primaryGoal'];
            if (!\is_string($goal) || !\in_array($goal, AthleteProfile::GOALS, true)) {
                throw new InvalidAthleteProfileException('primaryGoal is invalid');
            }
            $profile->setPrimaryGoal($goal);
        }

        $zoneMethod = $payload['zoneMethod'] ?? $profile->getZoneMethod();
        if (!\is_string($zoneMethod) || !\in_array($zoneMethod, AthleteProfile::ZONE_METHODS, true)) {
            throw new InvalidAthleteProfileException('zoneMethod is invalid');
        }

        $hrMax = $this->resolveHrMax($payload, $profile, $birthYear ?? $profile->getBirthYear());
        $hrRest = $this->optionalInt($payload, 'hrRest');
        if (\array_key_exists('hrRest', $payload) && null === $payload['hrRest']) {
            $hrRest = null;
        } elseif (!\array_key_exists('hrRest', $payload)) {
            $hrRest = $profile->getHrRest();
        }

        if ($hrMax < 120 || $hrMax > 230) {
            throw new InvalidAthleteProfileException('hrMax must be between 120 and 230');
        }
        if (null !== $hrRest && ($hrRest < 30 || $hrRest > 120 || $hrRest >= $hrMax)) {
            throw new InvalidAthleteProfileException('hrRest must be between 30 and 120 and below hrMax');
        }

        if (AthleteProfile::ZONE_METHOD_HRR_PCT === $zoneMethod && null === $hrRest) {
            throw new InvalidAthleteProfileException('hrRest is required for heart-rate reserve zones');
        }

        $hrMaxSource = $payload['hrMaxSource'] ?? $profile->getHrMaxSource();
        if (!\is_string($hrMaxSource) || !\in_array($hrMaxSource, [
            AthleteProfile::HR_MAX_SOURCE_MANUAL,
            AthleteProfile::HR_MAX_SOURCE_ESTIMATED_AGE,
        ], true)) {
            throw new InvalidAthleteProfileException('hrMaxSource is invalid');
        }

        $zones = $this->resolveZones($payload, $zoneMethod, $hrMax, $hrRest);

        $profile
            ->setHrMax($hrMax)
            ->setHrRest($hrRest)
            ->setHrMaxSource($hrMaxSource)
            ->setZoneMethod($zoneMethod)
            ->setZones($zones);

        if ($completeOnboarding && !$profile->isOnboardingCompleted()) {
            $profile->setOnboardingCompletedAt(new \DateTimeImmutable());
        }

        $this->entityManager->flush();

        return $profile;
    }

    /**
     * @param array<string, mixed> $payload
     *
     * @return list<array{zone: int, minBpm: int, maxBpm: int}>
     */
    private function resolveZones(array $payload, string $zoneMethod, int $hrMax, ?int $hrRest): array
    {
        if (AthleteProfile::ZONE_METHOD_CUSTOM === $zoneMethod) {
            return $this->parseCustomZones($payload['zones'] ?? null);
        }

        if (AthleteProfile::ZONE_METHOD_HRR_PCT === $zoneMethod) {
            return $this->hrZoneCalculator->fromHeartRateReserve($hrMax, (int) $hrRest);
        }

        return $this->hrZoneCalculator->fromPercentOfHrMax($hrMax);
    }

    /**
     * @return list<array{zone: int, minBpm: int, maxBpm: int}>
     */
    private function parseCustomZones(mixed $raw): array
    {
        if (!\is_array($raw) || 5 !== \count($raw)) {
            throw new InvalidAthleteProfileException('zones must contain exactly 5 entries');
        }

        $zones = [];
        foreach ($raw as $index => $row) {
            if (!\is_array($row)) {
                throw new InvalidAthleteProfileException('each zone must be an object');
            }
            $zone = isset($row['zone']) ? (int) $row['zone'] : $index + 1;
            $min = isset($row['minBpm']) ? (int) $row['minBpm'] : 0;
            $max = isset($row['maxBpm']) ? (int) $row['maxBpm'] : 0;
            if ($zone !== $index + 1) {
                throw new InvalidAthleteProfileException('zones must be ordered Z1–Z5');
            }
            if ($min < 40 || $max > 240 || $min > $max) {
                throw new InvalidAthleteProfileException(sprintf('zone %d bounds are invalid', $zone));
            }
            if ($index > 0 && $min < $zones[$index - 1]['maxBpm']) {
                throw new InvalidAthleteProfileException('zones must be non-overlapping and ascending');
            }
            $zones[] = ['zone' => $zone, 'minBpm' => $min, 'maxBpm' => $max];
        }

        return $zones;
    }

    /**
     * @param array<string, mixed> $payload
     */
    private function resolveHrMax(array $payload, AthleteProfile $profile, ?int $birthYear): int
    {
        if (\array_key_exists('hrMax', $payload) && null !== $payload['hrMax'] && '' !== $payload['hrMax']) {
            if (!is_numeric($payload['hrMax'])) {
                throw new InvalidAthleteProfileException('hrMax must be a number');
            }

            return (int) $payload['hrMax'];
        }

        if ($profile->getHrMax() > 0) {
            return $profile->getHrMax();
        }

        if (null !== $birthYear) {
            return $this->hrZoneCalculator->estimateHrMaxFromBirthYear($birthYear);
        }

        throw new InvalidAthleteProfileException('hrMax or birthYear is required');
    }

    /**
     * @param array<string, mixed> $payload
     */
    private function optionalInt(array $payload, string $key): ?int
    {
        if (!\array_key_exists($key, $payload) || null === $payload[$key] || '' === $payload[$key]) {
            return null;
        }
        if (!is_numeric($payload[$key])) {
            throw new InvalidAthleteProfileException(sprintf('%s must be a number', $key));
        }

        return (int) $payload[$key];
    }
}
