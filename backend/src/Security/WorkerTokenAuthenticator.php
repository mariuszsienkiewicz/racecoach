<?php

namespace App\Security;

use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Security\Core\Authentication\Token\TokenInterface;
use Symfony\Component\Security\Core\Exception\AuthenticationException;
use Symfony\Component\Security\Core\Exception\CustomUserMessageAuthenticationException;
use Symfony\Component\Security\Core\User\InMemoryUser;
use Symfony\Component\Security\Http\Authenticator\AbstractAuthenticator;
use Symfony\Component\Security\Http\Authenticator\Passport\Badge\UserBadge;
use Symfony\Component\Security\Http\Authenticator\Passport\Passport;
use Symfony\Component\Security\Http\Authenticator\Passport\SelfValidatingPassport;

/**
 * Authenticates internal workers via a shared secret header (X-Worker-Token).
 * Used instead of browser JWT for machine-to-machine calls.
 */
final class WorkerTokenAuthenticator extends AbstractAuthenticator
{
    public function __construct(
        private readonly string $workerApiToken,
    ) {
    }

    public function supports(Request $request): bool
    {
        $token = $request->headers->get('X-Worker-Token');

        return null !== $token && '' !== $token;
    }

    public function authenticate(Request $request): Passport
    {
        $provided = $request->headers->get('X-Worker-Token');
        if (null === $provided || '' === $provided || !hash_equals($this->workerApiToken, $provided)) {
            throw new CustomUserMessageAuthenticationException('Invalid worker token.');
        }

        // Loader callback, works on both worker_api and api (JWT) firewalls without
        // requiring the firewall's user provider to know about "worker".
        return new SelfValidatingPassport(new UserBadge(
            'worker',
            static fn (): InMemoryUser => new InMemoryUser('worker', null, ['ROLE_WORKER']),
        ));
    }

    public function onAuthenticationSuccess(Request $request, TokenInterface $token, string $firewallName): ?Response
    {
        return null;
    }

    public function onAuthenticationFailure(Request $request, AuthenticationException $exception): Response
    {
        return new JsonResponse(['message' => 'Unauthorized'], Response::HTTP_UNAUTHORIZED);
    }
}
