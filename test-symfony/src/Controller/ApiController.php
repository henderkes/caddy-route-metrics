<?php

namespace App\Controller;

use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\Routing\Attribute\Route;

class ApiController
{
    #[Route('/api/health', name: 'api_health', methods: ['GET'])]
    public function health(): JsonResponse
    {
        return new JsonResponse(['status' => 'ok']);
    }

    #[Route('/api/search', name: 'api_search', methods: ['GET'])]
    public function search(): JsonResponse
    {
        usleep(random_int(50000, 500000));
        return new JsonResponse(['results' => [], 'total' => 0]);
    }

    #[Route('/api/upload', name: 'api_upload', methods: ['POST'])]
    public function upload(): JsonResponse
    {
        usleep(random_int(100000, 400000));
        return new JsonResponse(['id' => random_int(1000, 9999), 'status' => 'uploaded'], 201);
    }

    #[Route('/api/reports/{id}', name: 'api_report_show', requirements: ['id' => '\d+'], methods: ['GET'])]
    public function report(int $id): JsonResponse
    {
        usleep(random_int(200000, 1000000));
        return new JsonResponse(['id' => $id, 'rows' => random_int(100, 10000), 'generated' => true]);
    }
}
