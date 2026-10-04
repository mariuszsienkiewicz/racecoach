<?php

namespace App\Service\Auth;

use App\Entity\User;
use App\Exception\EmailAlreadyRegisteredException;
use App\Exception\InvalidRegistrationException;
use App\Repository\UserRepository;
use Doctrine\DBAL\Exception\UniqueConstraintViolationException;
use Doctrine\ORM\EntityManagerInterface;
use Symfony\Component\PasswordHasher\Hasher\UserPasswordHasherInterface;

final class UserRegistrationService
{
    private const MIN_PASSWORD_LENGTH = 8;
    private const MAX_PASSWORD_LENGTH = 4096;
    private const MAX_EMAIL_LENGTH = 180;

    public function __construct(
        private readonly EntityManagerInterface $entityManager,
        private readonly UserPasswordHasherInterface $passwordHasher,
        private readonly UserRepository $userRepository,
    ) {
    }

    public function register(string $email, string $password): User
    {
        $email = $this->normalizeEmail($email);
        $this->assertValidEmail($email);
        $this->assertValidPassword($password);

        if ($this->userRepository->findOneBy(['email' => $email]) instanceof User) {
            throw new EmailAlreadyRegisteredException();
        }

        $user = new User();
        $user->setEmail($email);
        $user->setRoles(['ROLE_USER']);
        $user->setPassword($this->passwordHasher->hashPassword($user, $password));

        $this->entityManager->persist($user);

        try {
            $this->entityManager->flush();
        } catch (UniqueConstraintViolationException) {
            throw new EmailAlreadyRegisteredException();
        }

        return $user;
    }

    private function normalizeEmail(string $email): string
    {
        return strtolower(trim($email));
    }

    private function assertValidEmail(string $email): void
    {
        if ('' === $email) {
            throw new InvalidRegistrationException('email is required');
        }

        if (mb_strlen($email) > self::MAX_EMAIL_LENGTH) {
            throw new InvalidRegistrationException('email is too long');
        }

        if (false === filter_var($email, \FILTER_VALIDATE_EMAIL)) {
            throw new InvalidRegistrationException('email is invalid');
        }
    }

    private function assertValidPassword(string $password): void
    {
        if ('' === $password) {
            throw new InvalidRegistrationException('password is required');
        }

        $length = mb_strlen($password);
        if ($length < self::MIN_PASSWORD_LENGTH) {
            throw new InvalidRegistrationException('password must be at least 8 characters');
        }

        if ($length > self::MAX_PASSWORD_LENGTH) {
            throw new InvalidRegistrationException('password is too long');
        }
    }
}
