<?php

namespace App\Controller;

use App\Exception\EmailAlreadyRegisteredException;
use App\Exception\InvalidRegistrationException;
use App\Service\Auth\UserRegistrationService;
use Lexik\Bundle\JWTAuthenticationBundle\Services\JWTTokenManagerInterface;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Attribute\Route;

final class AuthController extends AbstractController
{
    public function __construct(
        private readonly UserRegistrationService $userRegistrationService,
        private readonly JWTTokenManagerInterface $jwtTokenManager,
    ) {
    }

    #[Route('/api/register', name: 'api_register', methods: ['POST'])]
    public function register(Request $request): JsonResponse
    {
        $payload = json_decode($request->getContent(), true);
        if (!\is_array($payload)) {
            return $this->json(['message' => 'Invalid JSON'], Response::HTTP_BAD_REQUEST);
        }

        $email = $payload['email'] ?? null;
        $password = $payload['password'] ?? null;
        if (!\is_string($email) || !\is_string($password)) {
            return $this->json(['message' => 'email and password are required'], Response::HTTP_BAD_REQUEST);
        }

        try {
            $user = $this->userRegistrationService->register($email, $password);
        } catch (InvalidRegistrationException $exception) {
            return $this->json(['message' => $exception->getMessage()], Response::HTTP_BAD_REQUEST);
        } catch (EmailAlreadyRegisteredException $exception) {
            return $this->json(['message' => $exception->getMessage()], Response::HTTP_CONFLICT);
        }

        return $this->json([
            'token' => $this->jwtTokenManager->create($user),
        ], Response::HTTP_CREATED);
    }
}
