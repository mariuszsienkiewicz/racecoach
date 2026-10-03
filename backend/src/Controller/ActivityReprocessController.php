<?php

namespace App\Controller;

use App\Entity\Activity;
use App\Entity\User;
use App\Exception\ActivityCannotReprocessException;
use App\Repository\ActivityRepository;
use App\Service\Pipeline\ActivityApiPresenter;
use App\Service\Pipeline\ActivityReprocessService;
use App\Service\Pipeline\ReprocessMode;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Attribute\Route;
use Symfony\Component\Security\Http\Attribute\CurrentUser;
use Symfony\Component\Security\Http\Attribute\IsGranted;

#[IsGranted('ROLE_USER')]
final class ActivityReprocessController extends AbstractController
{
    public function __construct(
        private readonly ActivityRepository $activityRepository,
        private readonly ActivityReprocessService $activityReprocessService,
        private readonly ActivityApiPresenter $activityApiPresenter,
    ) {
    }

    #[Route(
        '/api/activities/{id}/reprocess',
        name: 'api_activities_reprocess',
        methods: ['POST'],
        requirements: ['id' => '\d+'],
    )]
    public function __invoke(
        int $id,
        Request $request,
        #[CurrentUser] ?User $user,
    ): JsonResponse {
        if (!$user instanceof User) {
            return $this->json(['message' => 'Unauthorized'], Response::HTTP_UNAUTHORIZED);
        }

        $activity = $this->findOwnedActivity($id, $user);
        if (null === $activity) {
            return $this->json(['message' => 'Activity not found'], Response::HTTP_NOT_FOUND);
        }

        try {
            $mode = $this->parseMode($request);
        } catch (\InvalidArgumentException $exception) {
            return $this->json(['message' => $exception->getMessage()], Response::HTTP_BAD_REQUEST);
        }

        try {
            $activity = $this->activityReprocessService->reprocess($activity, $mode);
        } catch (ActivityCannotReprocessException $exception) {
            return $this->json(['message' => $exception->getMessage()], Response::HTTP_CONFLICT);
        }

        return $this->json($this->activityApiPresenter->present($activity), Response::HTTP_ACCEPTED);
    }

    private function findOwnedActivity(int $id, User $user): ?Activity
    {
        $activity = $this->activityRepository->find($id);
        if (!$activity instanceof Activity || $activity->getUser()->getId() !== $user->getId()) {
            return null;
        }

        return $activity;
    }

    private function parseMode(Request $request): ReprocessMode
    {
        $raw = $request->getContent();
        if ('' === trim($raw)) {
            return ReprocessMode::FULL;
        }

        try {
            $payload = json_decode($raw, true, 512, \JSON_THROW_ON_ERROR);
        } catch (\JsonException) {
            throw new \InvalidArgumentException('Invalid JSON');
        }

        if (!\is_array($payload)) {
            throw new \InvalidArgumentException('Invalid JSON');
        }

        if (!\array_key_exists('mode', $payload) || null === $payload['mode'] || '' === $payload['mode']) {
            return ReprocessMode::FULL;
        }

        if (!\is_string($payload['mode'])) {
            throw new \InvalidArgumentException('mode must be a string');
        }

        $mode = ReprocessMode::tryFrom($payload['mode']);
        if (!$mode instanceof ReprocessMode) {
            throw new \InvalidArgumentException('mode must be full or from_structure');
        }

        return $mode;
    }
}
