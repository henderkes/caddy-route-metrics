<?php

namespace App\Controller;

use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Attribute\Route;

class ErrorController
{
    #[Route('/error/bad-request', name: 'error_bad_request', methods: ['GET'])]
    public function badRequest(): JsonResponse
    {
        usleep(random_int(1000, 5000));
        return new JsonResponse(['error' => 'bad request'], 400);
    }

    #[Route('/error/forbidden', name: 'error_forbidden', methods: ['GET'])]
    public function forbidden(): JsonResponse
    {
        usleep(random_int(1000, 5000));
        return new JsonResponse(['error' => 'forbidden'], 403);
    }

    #[Route('/error/not-found', name: 'error_not_found', methods: ['GET'])]
    public function notFound(): JsonResponse
    {
        usleep(random_int(1000, 5000));
        return new JsonResponse(['error' => 'not found'], 404);
    }

    #[Route('/error/validation', name: 'error_validation', methods: ['POST'])]
    public function validation(): JsonResponse
    {
        usleep(random_int(5000, 15000));
        return new JsonResponse(['error' => 'validation failed', 'fields' => ['email']], 422);
    }

    #[Route('/error/server', name: 'error_server', methods: ['GET'])]
    public function server(): Response
    {
        usleep(random_int(100000, 300000));
        return new Response('internal server error', 500);
    }

    #[Route('/error/unavailable', name: 'error_unavailable', methods: ['GET'])]
    public function unavailable(): Response
    {
        usleep(random_int(1000, 5000));
        return new Response('service unavailable', 503);
    }
}
