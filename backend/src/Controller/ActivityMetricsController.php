<?php

namespace App\Controller;

use App\Entity\Activity;
use App\Repository\ActivityRepository;
use App\Service\Pipeline\AnalysisVersions;
use Doctrine\ORM\EntityManagerInterface;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Attribute\Route;
use Symfony\Component\Security\Http\Attribute\IsGranted;

#[IsGranted('ROLE_WORKER')]
final class ActivityMetricsController extends AbstractController
{
    #[Route(
        '/api/activities/{id}/metrics',
        name: 'api_activities_metrics_update',
        methods: ['PATCH'],
        requirements: ['id' => '\d+'],
    )]
    public function updateMetrics(
        int $id,
        Request $request,
        ActivityRepository $activityRepository,
        EntityManagerInterface $entityManager,
    ): JsonResponse {
        $activity = $activityRepository->find($id);
        if (!$activity instanceof Activity) {
            return $this->json(['message' => 'Activity not found'], Response::HTTP_NOT_FOUND);
        }

        $payload = json_decode($request->getContent(), true);
        if (!\is_array($payload)) {
            return $this->json(['message' => 'Invalid JSON'], Response::HTTP_BAD_REQUEST);
        }

        foreach (['distanceM', 'durationSec', 'avgHeartRate'] as $required) {
            if (!\array_key_exists($required, $payload)) {
                return $this->json(['message' => sprintf('Missing field "%s"', $required)], Response::HTTP_BAD_REQUEST);
            }
        }

        $distanceM = (int) $payload['distanceM'];
        $durationSec = (int) $payload['durationSec'];
        $avgHeartRate = (int) $payload['avgHeartRate'];

        $activity->setDistanceM($distanceM);
        $activity->setDurationSec($durationSec);
        $activity->setAvgHeartRate($avgHeartRate);
        $activity->setSummary('Metrics ready, further analysis still running.');
        $activity->setStatus(Activity::STATUS_ANALYZING);
        $activity->touchPipelineHeartbeat();

        if ($distanceM > 0) {
            $activity->setAvgPaceSecPerKm((int) round($durationSec / ($distanceM / 1000)));
        }

        if (\array_key_exists('metricsObjectKey', $payload)) {
            $key = $payload['metricsObjectKey'];
            $activity->setMetricsObjectKey(\is_string($key) ? $key : null);
        }

        $entityManager->flush();

        return $this->json($activity->toApiArray());
    }

    #[Route(
        '/api/activities/{id}/structure',
        name: 'api_activities_structure_update',
        methods: ['PATCH'],
        requirements: ['id' => '\d+'],
    )]
    public function updateStructure(
        int $id,
        Request $request,
        ActivityRepository $activityRepository,
        EntityManagerInterface $entityManager,
    ): JsonResponse {
        $activity = $activityRepository->find($id);
        if (!$activity instanceof Activity) {
            return $this->json(['message' => 'Activity not found'], Response::HTTP_NOT_FOUND);
        }

        $payload = json_decode($request->getContent(), true);
        if (!\is_array($payload)) {
            return $this->json(['message' => 'Invalid JSON'], Response::HTTP_BAD_REQUEST);
        }

        foreach (['structureObjectKey'] as $required) {
            if (!\array_key_exists($required, $payload)) {
                return $this->json(['message' => sprintf('Missing field "%s"', $required)], Response::HTTP_BAD_REQUEST);
            }
        }

        $activity->setStatus(Activity::STATUS_ANALYZING);
        $activity->setStructureObjectKey($payload['structureObjectKey']);
        $activity->touchPipelineHeartbeat();

        $entityManager->flush();

        return $this->json($activity->toApiArray());
    }

    #[Route(
        '/api/activities/{id}/features',
        name: 'api_activities_features_update',
        methods: ['PATCH'],
        requirements: ['id' => '\d+'],
    )]
    public function updateFeatures(
        int $id,
        Request $request,
        ActivityRepository $activityRepository,
        EntityManagerInterface $entityManager,
    ): JsonResponse {
        $activity = $activityRepository->find($id);
        if (!$activity instanceof Activity) {
            return $this->json(['message' => 'Activity not found'], Response::HTTP_NOT_FOUND);
        }

        $payload = json_decode($request->getContent(), true);
        if (!\is_array($payload)) {
            return $this->json(['message' => 'Invalid JSON'], Response::HTTP_BAD_REQUEST);
        }

        foreach (['featuresObjectKey'] as $required) {
            if (!\array_key_exists($required, $payload)) {
                return $this->json(['message' => sprintf('Missing field "%s"', $required)], Response::HTTP_BAD_REQUEST);
            }
        }

        $activity->setStatus(Activity::STATUS_ANALYZING);
        $activity->setFeaturesObjectKey($payload['featuresObjectKey']);
        $activity->touchPipelineHeartbeat();

        $entityManager->flush();

        return $this->json($activity->toApiArray());
    }

    #[Route(
        '/api/activities/{id}/summary',
        name: 'api_activities_summary_update',
        methods: ['PATCH'],
        requirements: ['id' => '\d+'],
    )]
    public function updateSummary(
        int $id,
        Request $request,
        ActivityRepository $activityRepository,
        EntityManagerInterface $entityManager,
    ): JsonResponse {
        $activity = $activityRepository->find($id);
        if (!$activity instanceof Activity) {
            return $this->json(['message' => 'Activity not found'], Response::HTTP_NOT_FOUND);
        }

        $payload = json_decode($request->getContent(), true);
        if (!\is_array($payload)) {
            return $this->json(['message' => 'Invalid JSON'], Response::HTTP_BAD_REQUEST);
        }

        foreach (['summaryObjectKey', 'summary'] as $required) {
            if (!\array_key_exists($required, $payload)) {
                return $this->json(['message' => sprintf('Missing field "%s"', $required)], Response::HTTP_BAD_REQUEST);
            }
        }

        if (!\is_string($payload['summaryObjectKey']) || '' === $payload['summaryObjectKey']) {
            return $this->json(['message' => 'Invalid summaryObjectKey'], Response::HTTP_BAD_REQUEST);
        }
        if (!\is_string($payload['summary'])) {
            return $this->json(['message' => 'Invalid summary'], Response::HTTP_BAD_REQUEST);
        }

        $activity->setSummaryObjectKey($payload['summaryObjectKey']);
        $activity->setSummary($payload['summary']);
        $activity->setStatus(Activity::STATUS_READY);
        $activity->setAnalysisVersion(AnalysisVersions::CURRENT);
        $activity->clearPipelineHeartbeat();

        $entityManager->flush();

        return $this->json($activity->toApiArray());
    }
}
