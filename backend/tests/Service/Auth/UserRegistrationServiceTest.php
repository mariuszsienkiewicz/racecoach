<?php

namespace App\Tests\Service\Auth;

use App\Entity\User;
use App\Exception\EmailAlreadyRegisteredException;
use App\Exception\InvalidRegistrationException;
use App\Repository\UserRepository;
use App\Service\Auth\UserRegistrationService;
use Doctrine\ORM\EntityManagerInterface;
use PHPUnit\Framework\TestCase;
use Symfony\Component\PasswordHasher\Hasher\UserPasswordHasherInterface;

final class UserRegistrationServiceTest extends TestCase
{
    public function testRegisterPersistsNormalizedEmailAndHashedPassword(): void
    {
        $persisted = null;

        $em = $this->createMock(EntityManagerInterface::class);
        $em->expects(self::once())
            ->method('persist')
            ->willReturnCallback(static function (object $entity) use (&$persisted): void {
                $persisted = $entity;
            });
        $em->expects(self::once())->method('flush');

        $hasher = $this->createMock(UserPasswordHasherInterface::class);
        $hasher->expects(self::once())
            ->method('hashPassword')
            ->willReturnCallback(static function (User $user, string $plain): string {
                self::assertSame('coach@racecoach.local', $user->getEmail());
                self::assertSame('secret123', $plain);

                return 'hashed-secret';
            });

        $users = $this->createMock(UserRepository::class);
        $users->expects(self::once())
            ->method('findOneBy')
            ->with(['email' => 'coach@racecoach.local'])
            ->willReturn(null);

        $service = new UserRegistrationService($em, $hasher, $users);
        $user = $service->register('  Coach@RaceCoach.local ', 'secret123');

        self::assertSame($persisted, $user);
        self::assertSame('coach@racecoach.local', $user->getEmail());
        self::assertSame('hashed-secret', $user->getPassword());
        self::assertContains('ROLE_USER', $user->getRoles());
    }

    public function testDuplicateEmailThrowsConflict(): void
    {
        $users = $this->createMock(UserRepository::class);
        $users->expects(self::once())
            ->method('findOneBy')
            ->with(['email' => 'coach@racecoach.local'])
            ->willReturn(new User());

        $service = new UserRegistrationService(
            $this->createStub(EntityManagerInterface::class),
            $this->createStub(UserPasswordHasherInterface::class),
            $users,
        );

        $this->expectException(EmailAlreadyRegisteredException::class);
        $service->register('coach@racecoach.local', 'secret123');
    }

    public function testShortPasswordIsRejected(): void
    {
        $service = new UserRegistrationService(
            $this->createStub(EntityManagerInterface::class),
            $this->createStub(UserPasswordHasherInterface::class),
            $this->createStub(UserRepository::class),
        );

        $this->expectException(InvalidRegistrationException::class);
        $this->expectExceptionMessage('password must be at least 8 characters');
        $service->register('coach@racecoach.local', 'short');
    }
}
