<?php

namespace App\Entity;

use App\Repository\AthleteProfileRepository;
use Doctrine\DBAL\Types\Types;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity(repositoryClass: AthleteProfileRepository::class)]
#[ORM\Table(name: 'athlete_profile')]
class AthleteProfile
{
    public const ZONE_METHOD_HRMAX_PCT = 'hrmax_pct';
    public const ZONE_METHOD_HRR_PCT = 'hrr_pct';
    public const ZONE_METHOD_CUSTOM = 'custom';

    public const HR_MAX_SOURCE_MANUAL = 'manual';
    public const HR_MAX_SOURCE_ESTIMATED_AGE = 'estimated_age';

    public const SEX_MALE = 'male';
    public const SEX_FEMALE = 'female';
    public const SEX_OTHER = 'other';
    public const SEX_UNSPECIFIED = 'unspecified';

    public const GOAL_GENERAL = 'general';
    public const GOAL_5K = '5k';
    public const GOAL_10K = '10k';
    public const GOAL_HALF = 'half';
    public const GOAL_MARATHON = 'marathon';

    /** @var list<string> */
    public const ZONE_METHODS = [
        self::ZONE_METHOD_HRMAX_PCT,
        self::ZONE_METHOD_HRR_PCT,
        self::ZONE_METHOD_CUSTOM,
    ];

    /** @var list<string> */
    public const GOALS = [
        self::GOAL_GENERAL,
        self::GOAL_5K,
        self::GOAL_10K,
        self::GOAL_HALF,
        self::GOAL_MARATHON,
    ];

    /** @var list<string> */
    public const SEXES = [
        self::SEX_MALE,
        self::SEX_FEMALE,
        self::SEX_OTHER,
        self::SEX_UNSPECIFIED,
    ];

    #[ORM\Id]
    #[ORM\GeneratedValue]
    #[ORM\Column]
    private ?int $id = null;

    #[ORM\OneToOne(inversedBy: 'athleteProfile')]
    #[ORM\JoinColumn(nullable: false, onDelete: 'CASCADE', unique: true)]
    private User $user;

    #[ORM\Column(nullable: true)]
    private ?int $birthYear = null;

    #[ORM\Column(length: 32, nullable: true)]
    private ?string $sex = null;

    #[ORM\Column(length: 32)]
    private string $primaryGoal = self::GOAL_GENERAL;

    #[ORM\Column]
    private int $hrMax = 0;

    #[ORM\Column(nullable: true)]
    private ?int $hrRest = null;

    #[ORM\Column(length: 32)]
    private string $hrMaxSource = self::HR_MAX_SOURCE_MANUAL;

    #[ORM\Column(length: 32)]
    private string $zoneMethod = self::ZONE_METHOD_HRMAX_PCT;

    /**
     * Materialized Z1–Z5 bounds: list of {zone, minBpm, maxBpm}.
     *
     * @var list<array{zone: int, minBpm: int, maxBpm: int}>
     */
    #[ORM\Column(type: Types::JSON)]
    private array $zones = [];

    #[ORM\Column]
    private \DateTimeImmutable $zonesUpdatedAt;

    #[ORM\Column(nullable: true)]
    private ?\DateTimeImmutable $onboardingCompletedAt = null;

    public function __construct()
    {
        $this->zonesUpdatedAt = new \DateTimeImmutable();
    }

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getUser(): User
    {
        return $this->user;
    }

    public function setUser(User $user): static
    {
        $this->user = $user;

        return $this;
    }

    public function getBirthYear(): ?int
    {
        return $this->birthYear;
    }

    public function setBirthYear(?int $birthYear): static
    {
        $this->birthYear = $birthYear;

        return $this;
    }

    public function getSex(): ?string
    {
        return $this->sex;
    }

    public function setSex(?string $sex): static
    {
        $this->sex = $sex;

        return $this;
    }

    public function getPrimaryGoal(): string
    {
        return $this->primaryGoal;
    }

    public function setPrimaryGoal(string $primaryGoal): static
    {
        $this->primaryGoal = $primaryGoal;

        return $this;
    }

    public function getHrMax(): int
    {
        return $this->hrMax;
    }

    public function setHrMax(int $hrMax): static
    {
        $this->hrMax = $hrMax;

        return $this;
    }

    public function getHrRest(): ?int
    {
        return $this->hrRest;
    }

    public function setHrRest(?int $hrRest): static
    {
        $this->hrRest = $hrRest;

        return $this;
    }

    public function getHrMaxSource(): string
    {
        return $this->hrMaxSource;
    }

    public function setHrMaxSource(string $hrMaxSource): static
    {
        $this->hrMaxSource = $hrMaxSource;

        return $this;
    }

    public function getZoneMethod(): string
    {
        return $this->zoneMethod;
    }

    public function setZoneMethod(string $zoneMethod): static
    {
        $this->zoneMethod = $zoneMethod;

        return $this;
    }

    /**
     * @return list<array{zone: int, minBpm: int, maxBpm: int}>
     */
    public function getZones(): array
    {
        return $this->zones;
    }

    /**
     * @param list<array{zone: int, minBpm: int, maxBpm: int}> $zones
     */
    public function setZones(array $zones): static
    {
        $this->zones = $zones;
        $this->zonesUpdatedAt = new \DateTimeImmutable();

        return $this;
    }

    public function getZonesUpdatedAt(): \DateTimeImmutable
    {
        return $this->zonesUpdatedAt;
    }

    public function getOnboardingCompletedAt(): ?\DateTimeImmutable
    {
        return $this->onboardingCompletedAt;
    }

    public function setOnboardingCompletedAt(?\DateTimeImmutable $onboardingCompletedAt): static
    {
        $this->onboardingCompletedAt = $onboardingCompletedAt;

        return $this;
    }

    public function isOnboardingCompleted(): bool
    {
        return null !== $this->onboardingCompletedAt;
    }

    /**
     * @return array<string, mixed>
     */
    public function toApiArray(): array
    {
        return [
            'birthYear' => $this->birthYear,
            'sex' => $this->sex,
            'primaryGoal' => $this->primaryGoal,
            'hrMax' => $this->hrMax,
            'hrRest' => $this->hrRest,
            'hrMaxSource' => $this->hrMaxSource,
            'zoneMethod' => $this->zoneMethod,
            'zones' => $this->zones,
            'zonesUpdatedAt' => $this->zonesUpdatedAt->format(\DateTimeInterface::ATOM),
            'onboardingCompleted' => $this->isOnboardingCompleted(),
            'onboardingCompletedAt' => $this->onboardingCompletedAt?->format(\DateTimeInterface::ATOM),
        ];
    }

    /**
     * @return array<string, mixed>
     */
    public function toAthleteContextArray(): array
    {
        return [
            'hrMax' => $this->hrMax,
            'hrRest' => $this->hrRest,
            'zoneMethod' => $this->zoneMethod,
            'primaryGoal' => $this->primaryGoal,
            'zones' => $this->zones,
            'zonesUpdatedAt' => $this->zonesUpdatedAt->format(\DateTimeInterface::ATOM),
        ];
    }
}
