<?php

namespace App\Repository;

use App\Entity\AthleteProfile;
use App\Entity\User;
use Doctrine\Bundle\DoctrineBundle\Repository\ServiceEntityRepository;
use Doctrine\Persistence\ManagerRegistry;

/**
 * @extends ServiceEntityRepository<AthleteProfile>
 */
class AthleteProfileRepository extends ServiceEntityRepository
{
    public function __construct(ManagerRegistry $registry)
    {
        parent::__construct($registry, AthleteProfile::class);
    }

    public function findOneByUser(User $user): ?AthleteProfile
    {
        return $this->findOneBy(['user' => $user]);
    }
}
