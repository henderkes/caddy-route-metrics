<?php

namespace App\Controller;

use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\Routing\Attribute\Route;

class OrderController
{
    #[Route('/orders/{uuid}', name: 'order_show', requirements: ['uuid' => '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'], methods: ['GET'])]
    public function show(string $uuid): JsonResponse
    {
        usleep(random_int(15000, 80000));
        return new JsonResponse(['uuid' => $uuid, 'status' => 'shipped', 'total' => 49.99]);
    }

    #[Route('/orders/{uuid}/cancel', name: 'order_cancel', requirements: ['uuid' => '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'], methods: ['PUT'])]
    public function cancel(string $uuid): JsonResponse
    {
        usleep(random_int(50000, 200000));
        return new JsonResponse(['uuid' => $uuid, 'status' => 'cancelled']);
    }
}
