<?php

namespace App\Tests\Service\Chat;

use App\Exception\Chat\FitChatProtocolException;
use App\Exception\Chat\FitChatRejectedException;
use App\Exception\Chat\FitChatUnavailableException;
use App\Service\Chat\FitChatAskRequest;
use App\Service\Chat\FitChatClient;
use PHPUnit\Framework\TestCase;
use Psr\Log\NullLogger;
use Symfony\Component\HttpClient\MockHttpClient;
use Symfony\Component\HttpClient\Response\MockResponse;

final class FitChatClientTest extends TestCase
{
    private function sampleAsk(): FitChatAskRequest
    {
        return new FitChatAskRequest(
            userId: 1,
            activityId: 104,
            message: 'How was the pace?',
            storageBucket: 'racecoach-fits',
            featuresObjectKey: 'users/1/fits/x.features.json',
            summaryObjectKey: 'users/1/fits/x.summary.json',
            summary: 'Solid session.',
        );
    }

    private function sseBody(string ...$chunks): string
    {
        return implode('', $chunks);
    }

    public function testAskStreamReturnsReplyAndTokens(): void
    {
        $body = $this->sseBody(
            "event: token\ndata: {\"text\":\"Nice\"}\n\n",
            "event: token\ndata: {\"text\":\" pacing.\"}\n\n",
            "event: done\ndata: {\"reply\":\"Nice pacing.\"}\n\n",
        );

        $http = new MockHttpClient([
            new MockResponse($body, ['http_code' => 200]),
        ]);

        $client = new FitChatClient($http, 'http://fit-chat:8081', 'secret', new NullLogger());
        $tokens = [];
        $reply = $client->askStream($this->sampleAsk(), static function (string $text) use (&$tokens): void {
            $tokens[] = $text;
        });

        self::assertSame('Nice pacing.', $reply->text);
        self::assertSame(['Nice', ' pacing.'], $tokens);
    }

    public function testAskStreamMapsServerErrorToUnavailable(): void
    {
        $http = new MockHttpClient([
            new MockResponse('{"message":"boom"}', ['http_code' => 503]),
        ]);
        $client = new FitChatClient($http, 'http://fit-chat:8081', 'secret', new NullLogger());

        $this->expectException(FitChatUnavailableException::class);
        $client->askStream($this->sampleAsk(), static fn (string $text) => null);
    }

    public function testAskStreamMapsClientErrorToRejected(): void
    {
        $http = new MockHttpClient([
            new MockResponse('{"message":"message is required"}', ['http_code' => 400]),
        ]);
        $client = new FitChatClient($http, 'http://fit-chat:8081', 'secret', new NullLogger());

        try {
            $client->askStream(
                new FitChatAskRequest(
                    userId: 1,
                    activityId: 104,
                    message: 'Hi',
                    storageBucket: 'racecoach-fits',
                    featuresObjectKey: 'users/1/fits/x.features.json',
                ),
                static fn (string $text) => null,
            );
            self::fail('Expected FitChatRejectedException');
        } catch (FitChatRejectedException $exception) {
            self::assertSame(400, $exception->statusCode);
            self::assertSame('message is required', $exception->getMessage());
        }
    }

    public function testAskStreamMapsErrorEventToUnavailable(): void
    {
        $body = $this->sseBody(
            "event: token\ndata: {\"text\":\"Hi\"}\n\n",
            "event: error\ndata: {\"message\":\"llm down\"}\n\n",
        );
        $http = new MockHttpClient([
            new MockResponse($body, ['http_code' => 200]),
        ]);
        $client = new FitChatClient($http, 'http://fit-chat:8081', 'secret', new NullLogger());

        $this->expectException(FitChatUnavailableException::class);
        $this->expectExceptionMessage('llm down');
        $client->askStream($this->sampleAsk(), static fn (string $text) => null);
    }

    public function testAskStreamMissingDoneIsProtocolError(): void
    {
        $body = $this->sseBody(
            "event: token\ndata: {\"text\":\"Hi\"}\n\n",
        );
        $http = new MockHttpClient([
            new MockResponse($body, ['http_code' => 200]),
        ]);
        $client = new FitChatClient($http, 'http://fit-chat:8081', 'secret', new NullLogger());

        $this->expectException(FitChatProtocolException::class);
        $client->askStream($this->sampleAsk(), static fn (string $text) => null);
    }

    public function testAskStreamSendsBearerAndArtifactKeys(): void
    {
        $http = new MockHttpClient(function (string $method, string $url, array $options): MockResponse {
            self::assertSame('POST', $method);
            self::assertSame('http://fit-chat:8081/v1/chat', $url);

            $authorization = $options['normalized_headers']['authorization'][0]
                ?? $options['headers']['Authorization']
                ?? null;
            self::assertSame('Authorization: Bearer secret', $authorization);

            $accept = $options['normalized_headers']['accept'][0]
                ?? $options['headers']['Accept']
                ?? null;
            self::assertNotFalse(stripos((string) $accept, 'text/event-stream'));

            $body = json_decode($options['body'], true, 512, JSON_THROW_ON_ERROR);
            self::assertSame([
                'userId' => 7,
                'activityId' => 104,
                'message' => 'Tell me more',
                'storageBucket' => 'racecoach-fits',
                'featuresObjectKey' => 'users/7/a.features.json',
                'history' => [
                    ['role' => 'athlete', 'content' => 'Earlier question'],
                    ['role' => 'coach', 'content' => 'Earlier answer'],
                ],
                'summaryObjectKey' => 'users/7/a.summary.json',
                'summary' => 'Headline note',
            ], $body);

            return new MockResponse(
                "event: done\ndata: {\"reply\":\"ok\"}\n\n",
                ['http_code' => 200],
            );
        });

        $client = new FitChatClient($http, 'http://fit-chat:8081/', 'secret', new NullLogger());
        $client->askStream(new FitChatAskRequest(
            userId: 7,
            activityId: 104,
            message: 'Tell me more',
            storageBucket: 'racecoach-fits',
            featuresObjectKey: 'users/7/a.features.json',
            summaryObjectKey: 'users/7/a.summary.json',
            summary: 'Headline note',
            history: [
                ['role' => 'athlete', 'content' => 'Earlier question'],
                ['role' => 'coach', 'content' => 'Earlier answer'],
            ],
        ), static fn (string $text) => null);
    }

    public function testAskStreamOmitsHistoryWhenNull(): void
    {
        $http = new MockHttpClient(function (string $method, string $url, array $options): MockResponse {
            $body = json_decode($options['body'], true, 512, JSON_THROW_ON_ERROR);
            self::assertArrayNotHasKey('history', $body);

            return new MockResponse(
                "event: done\ndata: {\"reply\":\"ok\"}\n\n",
                ['http_code' => 200],
            );
        });

        $client = new FitChatClient($http, 'http://fit-chat:8081/', 'secret', new NullLogger());
        $client->askStream($this->sampleAsk(), static fn (string $text) => null);
    }
}
