<?php

namespace App\Controller;

use Symfony\Component\HttpFoundation\RedirectResponse;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Attribute\Route;

class PageController
{
    #[Route('/', name: 'page_home')]
    public function home(): Response
    {
        usleep(random_int(1000, 5000));
        return new Response('<h1>Home</h1>');
    }

    #[Route('/about', name: 'page_about')]
    public function about(): Response
    {
        usleep(random_int(1000, 5000));
        return new Response('<h1>About</h1>');
    }

    #[Route('/contact', name: 'page_contact', methods: ['GET'])]
    public function contact(): Response
    {
        usleep(random_int(1000, 5000));
        return new Response('<h1>Contact</h1>');
    }

    #[Route('/contact', name: 'page_contact_submit', methods: ['POST'])]
    public function contactSubmit(): RedirectResponse
    {
        usleep(random_int(10000, 30000));
        return new RedirectResponse('/contact', 302);
    }

    #[Route('/login', name: 'page_login', methods: ['GET'])]
    public function login(): Response
    {
        usleep(random_int(1000, 5000));
        return new Response('<h1>Login</h1>');
    }

    #[Route('/login', name: 'page_login_submit', methods: ['POST'])]
    public function loginSubmit(): RedirectResponse
    {
        usleep(random_int(10000, 50000));
        return new RedirectResponse('/', 301);
    }
}
