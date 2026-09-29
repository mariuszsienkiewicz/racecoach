<?php

namespace App\Repository;

use App\Entity\Activity;
use App\Entity\ActivityChatMessage;
use Doctrine\Bundle\DoctrineBundle\Repository\ServiceEntityRepository;
use Doctrine\Persistence\ManagerRegistry;

/**
 * @extends ServiceEntityRepository<ActivityChatMessage>
 */
class ActivityChatMessageRepository extends ServiceEntityRepository
{
    public function __construct(ManagerRegistry $registry)
    {
        parent::__construct($registry, ActivityChatMessage::class);
    }

    /**
     * @return list<ActivityChatMessage>
     */
    public function findByActivity(Activity $activity): array
    {
        /** @var list<ActivityChatMessage> $messages */
        $messages = $this->createQueryBuilder('am')
            ->where('am.activity = :activity')
            ->setParameter('activity', $activity)
            ->orderBy('am.createdAt', 'ASC')
            ->addOrderBy('am.id', 'ASC')
            ->getQuery()
            ->getResult();

        return $messages;
    }

    /**
     * @return list<ActivityChatMessage>
     */
    public function findRecentByActivity(Activity $activity, int $limit): array
    {
        if ($limit <= 0) {
            return [];
        }

        /** @var list<ActivityChatMessage> $newestFirst */
        $newestFirst = $this->createQueryBuilder('am')
            ->where('am.activity = :activity')
            ->setParameter('activity', $activity)
            ->orderBy('am.createdAt', 'DESC')
            ->addOrderBy('am.id', 'DESC')
            ->setMaxResults($limit)
            ->getQuery()
            ->getResult();

        return array_reverse($newestFirst);
    }
}
