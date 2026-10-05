<?php

namespace App\Controller;

use App\Entity\User;
use App\Exception\InvalidAthleteProfileException;
use App\Service\Athlete\AthleteProfileService;
use App\Service\Athlete\HrZoneCalculator;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Attribute\Route;
use Symfony\Component\Security\Http\Attribute\CurrentUser;
use Symfony\Component\Security\Http\Attribute\IsGranted;

#[IsGranted('ROLE_USER')]
final class AthleteProfileController extends AbstractController
{
    public function __construct(
        private readonly AthleteProfileService $athleteProfileService,
        private readonly HrZoneCalculator $hrZoneCalculator,
    ) {
    }

    #[Route('/api/me/profile', name: 'api_me_profile_get', methods: ['GET'])]
    public function get(#[CurrentUser] ?User $user): JsonResponse
    {
        if (!$user instanceof User) {
            return $this->json(['message' => 'Unauthorized'], Response::HTTP_UNAUTHORIZED);
        }

        $profile = $this->athleteProfileService->getForUser($user);
        if (null === $profile) {
            return $this->json([
                'onboardingCompleted' => false,
                'profile' => null,
            ]);
        }

        return $this->json([
            'onboardingCompleted' => $profile->isOnboardingCompleted(),
            'profile' => $profile->toApiArray(),
        ]);
    }

    #[Route('/api/me/profile', name: 'api_me_profile_patch', methods: ['PATCH', 'PUT'])]
    public function upsert(Request $request, #[CurrentUser] ?User $user): JsonResponse
    {
        if (!$user instanceof User) {
            return $this->json(['message' => 'Unauthorized'], Response::HTTP_UNAUTHORIZED);
        }

        $payload = json_decode($request->getContent(), true);
        if (!\is_array($payload)) {
            return $this->json(['message' => 'Invalid JSON'], Response::HTTP_BAD_REQUEST);
        }

        $complete = !\array_key_exists('completeOnboarding', $payload) || (bool) $payload['completeOnboarding'];

        try {
            $profile = $this->athleteProfileService->upsert($user, $payload, $complete);
        } catch (InvalidAthleteProfileException $exception) {
            return $this->json(['message' => $exception->getMessage()], Response::HTTP_BAD_REQUEST);
        }

        return $this->json([
            'onboardingCompleted' => $profile->isOnboardingCompleted(),
            'profile' => $profile->toApiArray(),
        ]);
    }

    #[Route('/api/me/profile/preview-zones', name: 'api_me_profile_preview_zones', methods: ['POST'])]
    public function previewZones(Request $request, #[CurrentUser] ?User $user): JsonResponse
    {
        if (!$user instanceof User) {
            return $this->json(['message' => 'Unauthorized'], Response::HTTP_UNAUTHORIZED);
        }

        $payload = json_decode($request->getContent(), true);
        if (!\is_array($payload)) {
            return $this->json(['message' => 'Invalid JSON'], Response::HTTP_BAD_REQUEST);
        }

        $hrMax = isset($payload['hrMax']) ? (int) $payload['hrMax'] : 0;
        $birthYear = isset($payload['birthYear']) ? (int) $payload['birthYear'] : null;
        $hrRestRaw = $payload['hrRest'] ?? null;
        $hrRest = (null !== $hrRestRaw && '' !== $hrRestRaw) ? (int) $hrRestRaw : null;
        $method = \is_string($payload['zoneMethod'] ?? null) ? $payload['zoneMethod'] : 'hrmax_pct';

        if ($hrMax <= 0 && null !== $birthYear) {
            $hrMax = $this->hrZoneCalculator->estimateHrMaxFromBirthYear($birthYear);
        }
        if ($hrMax < 120 || $hrMax > 230) {
            return $this->json(['message' => 'hrMax or birthYear is required'], Response::HTTP_BAD_REQUEST);
        }

        $zones = 'hrr_pct' === $method && null !== $hrRest
            ? $this->hrZoneCalculator->fromHeartRateReserve($hrMax, $hrRest)
            : $this->hrZoneCalculator->fromPercentOfHrMax($hrMax);

        return $this->json([
            'hrMax' => $hrMax,
            'zones' => $zones,
        ]);
    }
}
