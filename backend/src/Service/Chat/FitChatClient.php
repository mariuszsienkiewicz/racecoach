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
 * uses Bearer CHAT_SERVICE_TOKEN to authenticate and POST /v1/chat to stream a coach reply.
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
     * Streams a coach reply. Invokes $onToken for each text delta.
     *
     * @param callable(string):void $onToken
     *
     * @throws FitChatUnavailableException
     * @throws FitChatRejectedException
     * @throws FitChatProtocolException
     * @throws \InvalidArgumentException
     */
    public function askStream(FitChatAskRequest $request, callable $onToken): FitChatReply
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
        $json = $this->buildRequestJson($request, $message);

        try {
            $response = $this->httpClient->request('POST', $url, [
                'headers' => [
                    'Authorization' => 'Bearer '.$this->serviceToken,
                    'Accept' => 'text/event-stream',
                ],
                'json' => $json,
                'timeout' => $this->timeoutSeconds,
                'buffer' => false,
            ]);

            $status = $response->getStatusCode();
            if ($status >= 400) {
                $this->throwForErrorStatus($response, $activityId, $status);
            }

            $reply = $this->consumeSse($response, $activityId, $onToken);

            return new FitChatReply(text: $reply);
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
    private function buildRequestJson(FitChatAskRequest $request, string $message): array
    {
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

        return $json;
    }

    /**
     * @param callable(string):void $onToken
     */
    private function consumeSse(ResponseInterface $response, int $activityId, callable $onToken): string
    {
        $buffer = '';
        $eventName = null;
        $dataLines = [];
        $reply = null;

        foreach ($this->httpClient->stream($response, $this->timeoutSeconds) as $chunk) {
            if ($chunk->isTimeout()) {
                continue;
            }

            $buffer .= $chunk->getContent();

            while (false !== ($pos = strpos($buffer, "\n"))) {
                $line = substr($buffer, 0, $pos);
                $buffer = substr($buffer, $pos + 1);
                $line = rtrim($line, "\r");

                if ($line === '') {
                    if ($eventName === null && $dataLines === []) {
                        continue;
                    }

                    $event = $eventName ?? 'message';
                    $data = implode("\n", $dataLines);
                    $eventName = null;
                    $dataLines = [];

                    $reply = $this->handleSseEvent($event, $data, $activityId, $onToken, $reply);
                    continue;
                }

                if (str_starts_with($line, ':')) {
                    continue;
                }
                if (str_starts_with($line, 'event:')) {
                    $eventName = trim(substr($line, \strlen('event:')));
                    continue;
                }
                if (str_starts_with($line, 'data:')) {
                    $dataLines[] = ltrim(substr($line, \strlen('data:')));
                }
            }
        }

        if ($eventName !== null || $dataLines !== []) {
            $reply = $this->handleSseEvent(
                $eventName ?? 'message',
                implode("\n", $dataLines),
                $activityId,
                $onToken,
                $reply,
            );
        }

        if ($reply === null || trim($reply) === '') {
            $this->logger->warning('fit-chat stream ended without done/reply', [
                'activityId' => $activityId,
            ]);

            throw new FitChatProtocolException('fit-chat stream ended without a reply.');
        }

        return $reply;
    }

    /**
     * @param callable(string):void $onToken
     */
    private function handleSseEvent(
        string $event,
        string $data,
        int $activityId,
        callable $onToken,
        ?string $reply,
    ): ?string {
        $payload = json_decode($data, true);
        if (!\is_array($payload)) {
            throw new FitChatProtocolException('fit-chat returned invalid SSE data.');
        }

        return match ($event) {
            'token' => $this->handleTokenEvent($payload, $onToken, $reply),
            'done' => $this->handleDoneEvent($payload, $activityId),
            'error' => throw $this->exceptionFromErrorEvent($payload, $activityId),
            default => $reply,
        };
    }

    /**
     * @param array<string, mixed> $payload
     * @param callable(string):void $onToken
     */
    private function handleTokenEvent(array $payload, callable $onToken, ?string $reply): ?string
    {
        $text = $payload['text'] ?? null;
        if (!\is_string($text) || $text === '') {
            return $reply;
        }
        $onToken($text);

        return $reply;
    }

    /**
     * @param array<string, mixed> $payload
     */
    private function handleDoneEvent(array $payload, int $activityId): string
    {
        $reply = $payload['reply'] ?? null;
        if (!\is_string($reply) || trim($reply) === '') {
            $this->logger->warning('fit-chat done event missing reply', [
                'activityId' => $activityId,
            ]);

            throw new FitChatProtocolException('fit-chat done event did not contain a reply.');
        }

        return $reply;
    }

    /**
     * @param array<string, mixed> $payload
     */
    private function exceptionFromErrorEvent(array $payload, int $activityId): FitChatException
    {
        $message = $payload['message'] ?? null;
        $detail = \is_string($message) && trim($message) !== ''
            ? trim($message)
            : 'Chat request failed in fit-chat.';

        $this->logger->warning('fit-chat stream error event', [
            'activityId' => $activityId,
            'message' => $detail,
        ]);

        return new FitChatUnavailableException($detail);
    }

    private function throwForErrorStatus(ResponseInterface $response, int $activityId, int $status): never
    {
        $payload = $this->tryDecodeJson($response, $activityId, $status);

        if ($status >= 500) {
            $this->logger->warning('fit-chat server error', [
                'activityId' => $activityId,
                'status' => $status,
            ]);

            throw new FitChatUnavailableException(sprintf('fit-chat failed with status %d.', $status));
        }

        $detail = $this->extractErrorMessage($payload) ?? 'Chat request rejected by fit-chat.';
        $this->logger->notice('fit-chat rejected request', [
            'activityId' => $activityId,
            'status' => $status,
        ]);

        throw new FitChatRejectedException($detail, $status);
    }

    /**
     * @return array<string, mixed>
     */
    private function tryDecodeJson(ResponseInterface $response, int $activityId, int $status): array
    {
        try {
            /** @var array<string, mixed> $payload */
            $payload = $response->toArray(false);

            return $payload;
        } catch (DecodingExceptionInterface|TransportExceptionInterface) {
            $this->logger->warning('fit-chat error body was not JSON', [
                'activityId' => $activityId,
                'status' => $status,
            ]);

            return [];
        }
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
