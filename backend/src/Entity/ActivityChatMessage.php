<?php

namespace App\Entity;

use App\Repository\ActivityChatMessageRepository;
use Doctrine\ORM\Mapping as ORM;
use Doctrine\DBAL\Types\Types;

#[ORM\Entity(repositoryClass: ActivityChatMessageRepository::class)]
#[ORM\Table(name: 'activity_chat_message')]
#[ORM\Index(columns: ['activity_id', 'created_at'], name: 'idx_activity_chat_message_activity_id_created_at')]
class ActivityChatMessage
{
    #[ORM\Id]
    #[ORM\GeneratedValue]
    #[ORM\Column]
    private ?int $id = null;

    #[ORM\ManyToOne(inversedBy: 'chatMessages')]
    #[ORM\JoinColumn(nullable: false, onDelete: 'CASCADE')]
    private Activity $activity;

    #[ORM\Column(type: Types::TEXT)]
    private string $content;

    #[ORM\Column(length: 16)]
    private string $role;

    #[ORM\Column(type: Types::DATETIME_IMMUTABLE)]
    private \DateTimeImmutable $createdAt;

    public function __construct()
    {
        $this->createdAt = new \DateTimeImmutable();
    }

    public function getId(): ?int
    {
        return $this->id;
    }
    
    public function getActivity(): Activity
    {
        return $this->activity;
    }

    public function setActivity(Activity $activity): static
    {
        $this->activity = $activity;

        return $this;
    }

    public function getContent(): string
    {
        return $this->content;
    }

    public function setContent(string $content): static
    {
        $this->content = $content;

        return $this;
    }

    public function getRole(): ActivityChatRole
    {
        return ActivityChatRole::from($this->role);
    }

    public function setRole(ActivityChatRole $role): static
    {
        $this->role = $role->value;

        return $this;
    }

    public function getCreatedAt(): \DateTimeImmutable
    {
        return $this->createdAt;
    }

    /**
     * @return array<string, mixed>
     */
    public function toApiArray(): array
    {
        return [
            'id' => (string) $this->id,
            'activityId' => (string) $this->activity->getId(),
            'content' => $this->content,
            'role' => $this->role,
            'createdAt' => $this->createdAt->format(\DateTimeInterface::ATOM),
        ];
    }
}
