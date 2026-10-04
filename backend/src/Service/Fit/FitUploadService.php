<?php

namespace App\Service\Fit;

use App\Entity\Activity;
use App\Entity\User;
use App\Exception\DuplicateFitUploadException;
use App\Message\ActivityUploadedMessage;
use App\Repository\ActivityRepository;
use App\Service\Storage\ObjectStorage;
use Doctrine\DBAL\Exception\UniqueConstraintViolationException;
use Doctrine\ORM\EntityManagerInterface;
use Symfony\Component\HttpFoundation\File\UploadedFile;
use Symfony\Component\Messenger\Bridge\Amqp\Transport\AmqpStamp;
use Symfony\Component\Messenger\MessageBusInterface;
use Symfony\Component\String\Slugger\SluggerInterface;

final class FitUploadService
{
    public function __construct(
        private readonly FitFileValidator $validator,
        private readonly ObjectStorage $objectStorage,
        private readonly EntityManagerInterface $entityManager,
        private readonly ActivityRepository $activityRepository,
        private readonly SluggerInterface $slugger,
        private readonly MessageBusInterface $messageBus,
    ) {
    }

    public function upload(User $user, UploadedFile $file): Activity
    {
        $this->validator->validate($file);

        $userId = $user->getId();
        if (null === $userId) {
            throw new \LogicException('Authenticated user must be persisted before uploading a .fit file.');
        }

        $originalFilename = $this->sanitizeOriginalFilename($file->getClientOriginalName());
        $sourcePath = $file->getPathname();
        $checksum = hash_file('sha256', $sourcePath);
        if (false === $checksum) {
            throw new \RuntimeException('Unable to checksum the uploaded .fit file.');
        }

        $existing = $this->activityRepository->findOneByUserAndChecksum($user, $checksum);
        if ($existing instanceof Activity) {
            throw new DuplicateFitUploadException(existingActivityId: (int) $existing->getId(), message: sprintf('This .fit file was already uploaded as activity #%d.', (int) $existing->getId()));
        }

        $objectKey = $this->buildObjectKey($userId);

        // Put bytes into S3 only after duplicate check, avoids orphan objects on re-upload.
        $this->objectStorage->putFile(
            objectKey: $objectKey,
            sourcePath: $sourcePath,
            contentType: 'application/octet-stream',
        );

        $fileSize = $file->getSize();
        if (false === $fileSize) {
            $fileSize = (int) filesize($sourcePath);
        }

        $activity = (new Activity())
            ->setUser($user)
            ->setTitle($this->titleFromFilename($originalFilename))
            ->setType(Activity::TYPE_EASY)
            ->setSource(Activity::SOURCE_FIT_UPLOAD)
            ->setStatus(Activity::STATUS_UPLOADED)
            ->setStartedAt(new \DateTimeImmutable())
            ->setOriginalFilename($originalFilename)
            ->setStoredPath(null)
            ->setObjectKey($objectKey)
            ->setStorageBucket($this->objectStorage->getBucket())
            ->setFileSizeBytes((int) $fileSize)
            ->setChecksumSha256($checksum)
            ->setSummary('Queued for analysis.')
            ->touchPipelineHeartbeat();

        $this->entityManager->persist($activity);

        try {
            $this->entityManager->flush();
        } catch (UniqueConstraintViolationException) {
            $this->entityManager->clear();
            $race = $this->activityRepository->findOneByUserAndChecksum($user, $checksum);
            throw new DuplicateFitUploadException(existingActivityId: (int) ($race?->getId() ?? 0), message: 'This .fit file was already uploaded.');
        }

        $activityId = $activity->getId();
        if (null === $activityId) {
            throw new \LogicException('Activity must have an id after flush.');
        }

        $this->messageBus->dispatch(
            new ActivityUploadedMessage(
                activityId: $activityId,
                userId: $userId,
                originalFilename: $activity->getOriginalFilename(),
                storageBucket: (string) $activity->getStorageBucket(),
                objectKey: (string) $activity->getObjectKey(),
                checksumSha256: $checksum,
                fileSizeBytes: $activity->getFileSizeBytes(),
                uploadedAt: $activity->getCreatedAt()->format(\DateTimeInterface::ATOM),
            ),
            [new AmqpStamp('activity.uploaded')],
        );

        return $activity;
    }

    private function buildObjectKey(int $userId): string
    {
        $now = new \DateTimeImmutable();

        return sprintf(
            'users/%d/fits/%s/%s.fit',
            $userId,
            $now->format('Y/m'),
            bin2hex(random_bytes(16)),
        );
    }

    private function sanitizeOriginalFilename(string $originalFilename): string
    {
        $basename = pathinfo($originalFilename, PATHINFO_FILENAME);
        $safe = strtolower($this->slugger->slug($basename)->toString());
        if ('' === $safe) {
            $safe = 'activity';
        }

        return substr($safe, 0, 160).'.fit';
    }

    private function titleFromFilename(string $originalFilename): string
    {
        $basename = pathinfo($originalFilename, PATHINFO_FILENAME);

        return '' !== $basename ? $basename : 'Uploaded activity';
    }
}
