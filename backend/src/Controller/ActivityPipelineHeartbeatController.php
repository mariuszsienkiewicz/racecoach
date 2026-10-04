<?php

namespace App\Controller;

use App\Entity\Activity;
use App\Repository\ActivityRepository;
use Doctrine\ORM\EntityManagerInterface;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Attribute\Route;
use Symfony\Component\Security\Http\Attribute\IsGranted;

#[IsGranted('ROLE_WORKER')]
final class ActivityPipelineHeartbeatController extends AbstractController
{
    #[Route(
        '/api/activities/{id}/pipeline-heartbeat',
        name: 'api_activities_pipeline_heartbeat',
        methods: ['PATCH'],
        requirements: ['id' => '\d+'],
    )]
    public function __invoke(
        int $id,
        ActivityRepository $activityRepository,
        EntityManagerInterface $entityManager,
    ): JsonResponse {
        $activity = $activityRepository->find($id);
        if (!$activity instanceof Activity) {
            return $this->json(['message' => 'Activity not found'], Response::HTTP_NOT_FOUND);
        }

        $activity->touchPipelineHeartbeat();
        if (Activity::STATUS_UPLOADED === $activity->getStatus()) {
            $activity->setStatus(Activity::STATUS_ANALYZING);
        }
        $entityManager->flush();

        return $this->json([
            'id' => (string) $activity->getId(),
            'status' => $activity->getStatus(),
            'pipelineHeartbeatAt' => $activity->getPipelineHeartbeatAt()?->format(\DateTimeInterface::ATOM),
        ]);
    }
}
