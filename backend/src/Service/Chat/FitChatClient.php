<?php

namespace App\Service\Chat;

use App\Exception\Chat\FitChatException;
use App\Exception\Chat\FitChatProtocolException;
use App\Exception\Chat\FitChatRejectedException;
use App\Exception\Chat\FitChatUnavailableException;
use Psr\Log\LoggerInterface;
use Symfony\Contracts\HttpClient\Exception\DecodingExceptionInterface;
use Symfony\Contracts\HttpClient\Exception\TransportExceptionInterface;
use Symfony\Contracts\HttpClient\HttpClientInterface;
use Symfony\Contracts\HttpClient\ResponseInterface;

/**
 * Internal HTTP client for the Go fit-chat service.
 *
 * Symfony remains the public gateway (JWT, ownership). This client only speaks
 * the private contract: Bearer CHAT_SERVICE_TOKEN → POST /v1/chat.
 */
final class FitChatClient
{
    private const CHAT_PATH = '/v1/chat';
    private const DEFAULT_TIMEOUT_SECONDS = 120.0;

    public function __construct(
        private readonly HttpClientInterface $httpClient,
        private readonly string $baseUrl,
        private readonly string $serviceToken,
        private readonly LoggerInterface $logger,
        private readonly float $timeoutSeconds = self::DEFAULT_TIMEOUT_SECONDS,
    ) {
    }

    /**
     * @throws FitChatUnavailableException when fit-chat is down or returns 5xx
     * @throws FitChatRejectedException    when fit-chat returns 4xx
     * @throws FitChatProtocolException    when the response body is not usable
     * @throws \InvalidArgumentException   when caller passes empty/invalid fields
     */
    public function ask(FitChatAskRequest $request): FitChatReply
    {
        if ($request->userId <= 0 || $request->activityId <= 0) {
            throw new \InvalidArgumentException('userId and activityId must be positive integers.');
        }

        $message = trim($request->message);
        if ($message === '') {
            throw new \InvalidArgumentException('message must not be empty.');
        }
        if (trim($request->storageBucket) === '' || trim($request->featuresObjectKey) === '') {
            throw new \InvalidArgumentException('storageBucket and featuresObjectKey are required.');
        }

        $url = rtrim($this->baseUrl, '/').self::CHAT_PATH;
        $activityId = $request->activityId;

        $json = [
            'userId' => $request->userId,
            'activityId' => $request->activityId,
            'message' => $message,
            'storageBucket' => $request->storageBucket,
            'featuresObjectKey' => $request->featuresObjectKey,
        ];
        if ($request->history !== null) {
            $json['history'] = $request->history;
        }
        if ($request->summaryObjectKey !== null && $request->summaryObjectKey !== '') {
            $json['summaryObjectKey'] = $request->summaryObjectKey;
        }
        if ($request->summary !== null && trim($request->summary) !== '') {
            $json['summary'] = $request->summary;
        }

        try {
            $response = $this->httpClient->request('POST', $url, [
                'headers' => [
                    'Authorization' => 'Bearer '.$this->serviceToken,
                    'Accept' => 'application/json',
                ],
                'json' => $json,
                'timeout' => $this->timeoutSeconds,
            ]);

            $status = $response->getStatusCode();
            $payload = $this->decodePayload($response, $activityId, $status);

            if ($status >= 500) {
                $this->logger->warning('fit-chat server error', [
                    'activityId' => $activityId,
                    'status' => $status,
                ]);

                throw new FitChatUnavailableException(sprintf('fit-chat failed with status %d.', $status));
            }

            if ($status >= 400) {
                $detail = $this->extractErrorMessage($payload) ?? 'Chat request rejected by fit-chat.';
                $this->logger->notice('fit-chat rejected request', [
                    'activityId' => $activityId,
                    'status' => $status,
                ]);

                throw new FitChatRejectedException($detail, $status);
            }

            return $this->mapReply($payload, $activityId, $status);
        } catch (FitChatException $exception) {
            throw $exception;
        } catch (TransportExceptionInterface $exception) {
            $this->logger->warning('fit-chat transport failure', [
                'activityId' => $activityId,
                'error' => $exception->getMessage(),
            ]);

            throw new FitChatUnavailableException('fit-chat is unreachable.', previous: $exception);
        }
    }

    /**
     * @return array<string, mixed>
     */
    private function decodePayload(ResponseInterface $response, int $activityId, int $status): array
    {
        try {
            /** @var array<string, mixed> $payload */
            $payload = $response->toArray(false);

            return $payload;
        } catch (DecodingExceptionInterface $exception) {
            $this->logger->warning('fit-chat returned non-JSON body', [
                'activityId' => $activityId,
                'status' => $status,
            ]);

            throw new FitChatProtocolException('fit-chat returned a non-JSON response.', previous: $exception);
        }
    }

    /**
     * @param array<string, mixed> $payload
     */
    private function mapReply(array $payload, int $activityId, int $status): FitChatReply
    {
        $reply = $payload['reply'] ?? null;
        if (!\is_string($reply) || trim($reply) === '') {
            $this->logger->warning('fit-chat response missing reply', [
                'activityId' => $activityId,
                'status' => $status,
            ]);

            throw new FitChatProtocolException('fit-chat response did not contain a reply.');
        }

        return new FitChatReply(text: $reply);
    }

    /**
     * @param array<string, mixed> $payload
     */
    private function extractErrorMessage(array $payload): ?string
    {
        foreach (['message', 'error'] as $key) {
            $value = $payload[$key] ?? null;
            if (\is_string($value) && trim($value) !== '') {
                return trim($value);
            }
        }

        return null;
    }
}
