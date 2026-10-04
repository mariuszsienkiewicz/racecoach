<?php

namespace App\Entity;

use App\Repository\ActivityRepository;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\DBAL\Types\Types;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity(repositoryClass: ActivityRepository::class)]
#[ORM\Table(name: 'activity')]
#[ORM\Index(columns: ['user_id', 'started_at'], name: 'idx_activity_user_started')]
#[ORM\HasLifecycleCallbacks]
class Activity
{
    public const SOURCE_FIT_UPLOAD = 'fit_upload';
    public const STATUS_UPLOADED = 'uploaded';
    public const STATUS_ANALYZING = 'analyzing';
    public const STATUS_READY = 'ready';
    public const STATUS_FAILED = 'failed';
    public const TYPE_EASY = 'easy';
    public const TYPE_TEMPO = 'tempo';
    public const TYPE_INTERVALS = 'intervals';
    public const TYPE_LONG_RUN = 'long_run';
    public const TYPE_RECOVERY = 'recovery';
    public const TYPE_RACE = 'race';

    /** @var list<string> */
    public const TYPES = [
        self::TYPE_EASY,
        self::TYPE_TEMPO,
        self::TYPE_INTERVALS,
        self::TYPE_LONG_RUN,
        self::TYPE_RECOVERY,
        self::TYPE_RACE,
    ];

    #[ORM\Id]
    #[ORM\GeneratedValue]
    #[ORM\Column]
    private ?int $id = null;

    #[ORM\ManyToOne]
    #[ORM\JoinColumn(nullable: false, onDelete: 'CASCADE')]
    private User $user;

    #[ORM\Column(length: 180)]
    private string $title = '';

    #[ORM\Column(length: 32)]
    private string $type = self::TYPE_EASY;

    #[ORM\Column(length: 32)]
    private string $source = self::SOURCE_FIT_UPLOAD;

    #[ORM\Column(length: 32)]
    private string $status = self::STATUS_UPLOADED;

    #[ORM\Column]
    private \DateTimeImmutable $startedAt;

    #[ORM\Column]
    private int $durationSec = 0;

    #[ORM\Column]
    private int $distanceM = 0;

    #[ORM\Column(nullable: true)]
    private ?int $avgPaceSecPerKm = null;

    #[ORM\Column(nullable: true)]
    private ?int $avgHeartRate = null;

    #[ORM\Column(nullable: true)]
    private ?int $elevationGainM = null;

    #[ORM\Column(type: Types::TEXT, nullable: true)]
    private ?string $summary = null;

    #[ORM\Column(length: 255)]
    private string $originalFilename = '';

    /**
     * Legacy local disk path from the pre-MinIO phase. Nullable for new uploads.
     */
    #[ORM\Column(length: 512, nullable: true)]
    private ?string $storedPath = null;

    /**
     * S3/MinIO object key, e.g. users/1/fits/abc123.fit.
     */
    #[ORM\Column(length: 512, nullable: true)]
    private ?string $objectKey = null;

    /**
     * S3/MinIO object key, e.g. users/1/abc123.metrics.json
     * If null, the metrics are not yet calculated.
     */
    #[ORM\Column(length: 512, nullable: true)]
    private ?string $metricsObjectKey = null;

    /**
     * S3/MinIO object key, e.g. users/1/abc123.structure.json
     * If null, the structure is not yet calculated.
     */
    #[ORM\Column(length: 512, nullable: true)]
    private ?string $structureObjectKey = null;

    /**
     * S3/MinIO object key, e.g. users/1/abc123.features.json
     * If null, the AI feature payload is not yet calculated.
     */
    #[ORM\Column(length: 512, nullable: true)]
    private ?string $featuresObjectKey = null;

    /**
     * S3/MinIO object key, e.g. users/1/abc123.summary.json
     * If null, the AI summary is not yet calculated.
     */
    #[ORM\Column(length: 512, nullable: true)]
    private ?string $summaryObjectKey = null;

    #[ORM\Column(length: 128, nullable: true)]
    private ?string $storageBucket = null;

    #[ORM\Column]
    private int $fileSizeBytes = 0;

    /** SHA-256 of the original .fit bytes, used to reject duplicate uploads per user. */
    #[ORM\Column(length: 64, nullable: true)]
    private ?string $checksumSha256 = null;

    #[ORM\Column(nullable: true)]
    private ?int $analysisVersion = null;

    #[ORM\Column]
    private \DateTimeImmutable $createdAt;

    #[ORM\Column]
    private \DateTimeImmutable $updatedAt;

    #[ORM\Column(nullable: true)]
    private ?\DateTimeImmutable $pipelineHeartbeatAt = null;

    /** @var Collection<int, ActivityChatMessage> */
    #[ORM\OneToMany(mappedBy: 'activity', targetEntity: ActivityChatMessage::class)]
    private Collection $chatMessages;

    public function __construct()
    {
        $now = new \DateTimeImmutable();
        $this->startedAt = $now;
        $this->createdAt = $now;
        $this->updatedAt = $now;
        $this->chatMessages = new ArrayCollection();
    }

    #[ORM\PrePersist]
    #[ORM\PreUpdate]
    public function touchUpdatedAt(): void
    {
        $this->updatedAt = new \DateTimeImmutable();
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

    public function getTitle(): string
    {
        return $this->title;
    }

    public function setTitle(string $title): static
    {
        $this->title = $title;

        return $this;
    }

    public function getType(): string
    {
        return $this->type;
    }

    public function setType(string $type): static
    {
        $this->type = $type;

        return $this;
    }

    public function getSource(): string
    {
        return $this->source;
    }

    public function setSource(string $source): static
    {
        $this->source = $source;

        return $this;
    }

    public function getStatus(): string
    {
        return $this->status;
    }

    public function setStatus(string $status): static
    {
        $this->status = $status;

        return $this;
    }

    public function getStartedAt(): \DateTimeImmutable
    {
        return $this->startedAt;
    }

    public function setStartedAt(\DateTimeImmutable $startedAt): static
    {
        $this->startedAt = $startedAt;

        return $this;
    }

    public function getDurationSec(): int
    {
        return $this->durationSec;
    }

    public function setDurationSec(int $durationSec): static
    {
        $this->durationSec = $durationSec;

        return $this;
    }

    public function getDistanceM(): int
    {
        return $this->distanceM;
    }

    public function setDistanceM(int $distanceM): static
    {
        $this->distanceM = $distanceM;

        return $this;
    }

    public function getAvgPaceSecPerKm(): ?int
    {
        return $this->avgPaceSecPerKm;
    }

    public function setAvgPaceSecPerKm(?int $avgPaceSecPerKm): static
    {
        $this->avgPaceSecPerKm = $avgPaceSecPerKm;

        return $this;
    }

    public function getAvgHeartRate(): ?int
    {
        return $this->avgHeartRate;
    }

    public function setAvgHeartRate(?int $avgHeartRate): static
    {
        $this->avgHeartRate = $avgHeartRate;

        return $this;
    }

    public function getElevationGainM(): ?int
    {
        return $this->elevationGainM;
    }

    public function setElevationGainM(?int $elevationGainM): static
    {
        $this->elevationGainM = $elevationGainM;

        return $this;
    }

    public function getSummary(): ?string
    {
        return $this->summary;
    }

    public function setSummary(?string $summary): static
    {
        $this->summary = $summary;

        return $this;
    }

    public function getOriginalFilename(): string
    {
        return $this->originalFilename;
    }

    public function setOriginalFilename(string $originalFilename): static
    {
        $this->originalFilename = $originalFilename;

        return $this;
    }

    public function getStoredPath(): ?string
    {
        return $this->storedPath;
    }

    public function setStoredPath(?string $storedPath): static
    {
        $this->storedPath = $storedPath;

        return $this;
    }

    public function getObjectKey(): ?string
    {
        return $this->objectKey;
    }

    public function setObjectKey(?string $objectKey): static
    {
        $this->objectKey = $objectKey;

        return $this;
    }

    public function getMetricsObjectKey(): ?string
    {
        return $this->metricsObjectKey;
    }

    public function setMetricsObjectKey(?string $metricsObjectKey): static
    {
        $this->metricsObjectKey = $metricsObjectKey;

        return $this;
    }

    public function getStructureObjectKey(): ?string
    {
        return $this->structureObjectKey;
    }

    public function setStructureObjectKey(?string $structureObjectKey): static
    {
        $this->structureObjectKey = $structureObjectKey;

        return $this;
    }

    public function getFeaturesObjectKey(): ?string
    {
        return $this->featuresObjectKey;
    }

    public function setFeaturesObjectKey(?string $featuresObjectKey): static
    {
        $this->featuresObjectKey = $featuresObjectKey;

        return $this;
    }

    public function getSummaryObjectKey(): ?string
    {
        return $this->summaryObjectKey;
    }

    public function setSummaryObjectKey(?string $summaryObjectKey): static
    {
        $this->summaryObjectKey = $summaryObjectKey;

        return $this;
    }

    public function getStorageBucket(): ?string
    {
        return $this->storageBucket;
    }

    public function setStorageBucket(?string $storageBucket): static
    {
        $this->storageBucket = $storageBucket;

        return $this;
    }

    public function getFileSizeBytes(): int
    {
        return $this->fileSizeBytes;
    }

    public function setFileSizeBytes(int $fileSizeBytes): static
    {
        $this->fileSizeBytes = $fileSizeBytes;

        return $this;
    }

    public function getChecksumSha256(): ?string
    {
        return $this->checksumSha256;
    }

    public function setChecksumSha256(?string $checksumSha256): static
    {
        $this->checksumSha256 = $checksumSha256;

        return $this;
    }

    public function getAnalysisVersion(): ?int
    {
        return $this->analysisVersion;
    }

    public function setAnalysisVersion(?int $analysisVersion): static
    {
        $this->analysisVersion = $analysisVersion;

        return $this;
    }

    public function getCreatedAt(): \DateTimeImmutable
    {
        return $this->createdAt;
    }

    public function getUpdatedAt(): \DateTimeImmutable
    {
        return $this->updatedAt;
    }

    public function getPipelineHeartbeatAt(): ?\DateTimeImmutable
    {
        return $this->pipelineHeartbeatAt;
    }

    public function touchPipelineHeartbeat(): static
    {
        $this->pipelineHeartbeatAt = new \DateTimeImmutable();

        return $this;
    }

    public function clearPipelineHeartbeat(): static
    {
        $this->pipelineHeartbeatAt = null;

        return $this;
    }

    /**
     * @return Collection<int, ActivityChatMessage>
     */
    public function getChatMessages(): Collection
    {
        return $this->chatMessages;
    }

    public function addChatMessage(ActivityChatMessage $chatMessage): static
    {
        $chatMessage->setActivity($this);
        $this->chatMessages->add($chatMessage);

        return $this;
    }

    /**
     * @return array<string, mixed>
     */
    public function toApiArray(): array
    {
        return [
            'id' => (string) $this->id,
            'title' => $this->title,
            'type' => $this->type,
            'source' => $this->source,
            'status' => $this->status,
            'startedAt' => $this->startedAt->format(\DateTimeInterface::ATOM),
            'durationSec' => $this->durationSec,
            'distanceM' => $this->distanceM,
            'avgPaceSecPerKm' => $this->avgPaceSecPerKm,
            'avgHeartRate' => $this->avgHeartRate,
            'elevationGainM' => $this->elevationGainM,
            'summary' => $this->summary,
            'metricsObjectKey' => $this->metricsObjectKey,
            'structureObjectKey' => $this->structureObjectKey,
            'featuresObjectKey' => $this->featuresObjectKey,
            'summaryObjectKey' => $this->summaryObjectKey,
            'checksumSha256' => $this->checksumSha256,
            'analysisVersion' => $this->analysisVersion,
            'pipelineHeartbeatAt' => $this->pipelineHeartbeatAt?->format(\DateTimeInterface::ATOM),
        ];
    }
}
