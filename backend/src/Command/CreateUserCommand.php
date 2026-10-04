<?php

namespace App\Command;

use App\Exception\EmailAlreadyRegisteredException;
use App\Exception\InvalidRegistrationException;
use App\Service\Auth\UserRegistrationService;
use Symfony\Component\Console\Attribute\AsCommand;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Input\InputArgument;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Output\OutputInterface;
use Symfony\Component\Console\Style\SymfonyStyle;

#[AsCommand(
    name: 'app:create-user',
    description: 'Create a user account for JWT login',
)]
final class CreateUserCommand extends Command
{
    public function __construct(
        private readonly UserRegistrationService $userRegistrationService,
    ) {
        parent::__construct();
    }

    protected function configure(): void
    {
        $this
            ->addArgument('email', InputArgument::REQUIRED)
            ->addArgument('password', InputArgument::REQUIRED);
    }

    protected function execute(InputInterface $input, OutputInterface $output): int
    {
        $io = new SymfonyStyle($input, $output);
        $email = (string) $input->getArgument('email');
        $password = (string) $input->getArgument('password');

        try {
            $user = $this->userRegistrationService->register($email, $password);
        } catch (EmailAlreadyRegisteredException $exception) {
            $io->error($exception->getMessage());

            return Command::FAILURE;
        } catch (InvalidRegistrationException $exception) {
            $io->error($exception->getMessage());

            return Command::FAILURE;
        }

        $io->success(sprintf('User "%s" created.', $user->getEmail()));

        return Command::SUCCESS;
    }
}
