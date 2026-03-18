<?php

namespace App\Controller;

use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Attribute\Route;

class UserController
{
    #[Route('/users', name: 'user_list', methods: ['GET'])]
    public function list(): JsonResponse
    {
        usleep(random_int(5000, 20000));
        return new JsonResponse([
            ['id' => 1, 'name' => 'Alice'],
            ['id' => 2, 'name' => 'Bob'],
        ]);
    }

    #[Route('/users', name: 'user_create', methods: ['POST'])]
    public function create(): JsonResponse
    {
        usleep(random_int(15000, 50000));
        return new JsonResponse(['id' => random_int(100, 999), 'name' => 'New User'], 201);
    }

    #[Route('/users/{id}', name: 'user_show', requirements: ['id' => '\d+'], methods: ['GET'])]
    public function show(int $id): JsonResponse
    {
        usleep(random_int(2000, 15000));
        return new JsonResponse(['id' => $id, 'name' => "User $id"]);
    }

    #[Route('/users/{id}', name: 'user_update', requirements: ['id' => '\d+'], methods: ['PUT'])]
    public function update(int $id): JsonResponse
    {
        usleep(random_int(10000, 40000));
        return new JsonResponse(['id' => $id, 'updated' => true]);
    }

    #[Route('/users/{id}', name: 'user_delete', requirements: ['id' => '\d+'], methods: ['DELETE'])]
    public function delete(int $id): Response
    {
        usleep(random_int(2000, 10000));
        return new Response('', 204);
    }

    #[Route('/users/{id}/posts', name: 'user_posts', requirements: ['id' => '\d+'], methods: ['GET'])]
    public function posts(int $id): JsonResponse
    {
        usleep(random_int(10000, 60000));
        return new JsonResponse([
            ['id' => 1, 'title' => "Post by user $id"],
            ['id' => 2, 'title' => "Another post by user $id"],
        ]);
    }

    #[Route('/users/{userId}/posts/{postId}', name: 'user_post_show', requirements: ['userId' => '\d+', 'postId' => '\d+'], methods: ['GET'])]
    public function postShow(int $userId, int $postId): JsonResponse
    {
        usleep(random_int(5000, 20000));
        return new JsonResponse(['userId' => $userId, 'postId' => $postId, 'title' => "Post $postId"]);
    }
}
