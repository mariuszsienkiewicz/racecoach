<?php

namespace App\Controller;

use App\Entity\User;
use App\Exception\DuplicateFitUploadException;
use App\Service\Fit\FitUploadService;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\File\UploadedFile;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\HttpKernel\Exception\BadRequestHttpException;
use Symfony\Component\Routing\Attribute\Route;
use Symfony\Component\Security\Http\Attribute\CurrentUser;
use Symfony\Component\Security\Http\Attribute\IsGranted;

#[IsGranted('ROLE_USER')]
final class ActivityUploadController extends AbstractController
{
    public function __construct(
        private readonly FitUploadService $fitUploadService,
    ) {
    }

    #[Route('/api/activities/fit', name: 'api_activities_fit_upload', methods: ['POST'])]
    public function __invoke(Request $request, #[CurrentUser] ?User $user): JsonResponse
    {
        if (!$user instanceof User) {
            return $this->json(['message' => 'Unauthorized'], Response::HTTP_UNAUTHORIZED);
        }

        $uploaded = $request->files->get('fitFile');
        if (!$uploaded instanceof UploadedFile) {
            throw new BadRequestHttpException('Missing multipart field "fitFile" with a .fit file.');
        }

        try {
            $activity = $this->fitUploadService->upload($user, $uploaded);
        } catch (DuplicateFitUploadException $exception) {
            return $this->json([
                'message' => $exception->getMessage(),
                'existingActivityId' => (string) $exception->existingActivityId,
            ], Response::HTTP_CONFLICT);
        }

        return $this->json($activity->toApiArray(), Response::HTTP_CREATED);
    }
}
