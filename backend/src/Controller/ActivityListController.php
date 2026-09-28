<?php

namespace App\Controller;

use App\Entity\Activity;
use App\Entity\User;
use App\Repository\ActivityRepository;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\ExpressionLanguage\Expression;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Attribute\Route;
use Symfony\Component\Security\Http\Attribute\CurrentUser;
use Symfony\Component\Security\Http\Attribute\IsGranted;

final class ActivityListController extends AbstractController
{
    #[IsGranted('ROLE_USER')]
    #[Route('/api/activities', name: 'api_activities_list', methods: ['GET'])]
    public function list(
        #[CurrentUser] ?User $user,
        ActivityRepository $activityRepository,
    ): JsonResponse {
        if (!$user instanceof User) {
            return $this->json(['message' => 'Unauthorized'], Response::HTTP_UNAUTHORIZED);
        }

        $activities = $activityRepository->findRecentForUser($user, 100);

        return $this->json([
            'activities' => array_map(
                static fn (Activity $activity) => $activity->toApiArray(),
                $activities,
            ),
        ]);
    }

    /**
     * Athletes use Bearer JWT.
     * Workers use X-Worker-Token.
     */
    #[IsGranted(new Expression("is_granted('ROLE_USER') or is_granted('ROLE_WORKER')"))]
    #[Route(
        '/api/activities/{id}',
        name: 'api_activity_get',
        methods: ['GET'],
        requirements: ['id' => '\d+'],
    )]
    public function getActivity(
        int $id,
        ActivityRepository $activityRepository,
    ): JsonResponse {
        $activity = $activityRepository->find($id);
        if (!$activity instanceof Activity) {
            return $this->json(['message' => 'Activity not found'], Response::HTTP_NOT_FOUND);
        }

        // Worker token, InMemoryUser with ROLE_WORKER
        if ($this->isGranted('ROLE_WORKER')) {
            return $this->json($activity->toApiArray());
        }

        $user = $this->getUser();
        if (!$user instanceof User || $activity->getUser()?->getId() !== $user->getId()) {
            return $this->json(['message' => 'Activity not found'], Response::HTTP_NOT_FOUND);
        }

        return $this->json($activity->toApiArray());
    }
}
