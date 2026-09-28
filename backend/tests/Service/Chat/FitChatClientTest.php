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

    public function testAskReturnsReplyOnSuccess(): void
    {
        $http = new MockHttpClient([
            new MockResponse(
                json_encode(['reply' => 'Nice pacing.'], JSON_THROW_ON_ERROR),
                ['http_code' => 200],
            ),
        ]);

        $client = new FitChatClient($http, 'http://fit-chat:8081', 'secret', new NullLogger());
        $reply = $client->ask($this->sampleAsk());

        self::assertSame('Nice pacing.', $reply->text);
    }

    public function testAskMapsServerErrorToUnavailable(): void
    {
        $http = new MockHttpClient([
            new MockResponse('{"message":"boom"}', ['http_code' => 503]),
        ]);
        $client = new FitChatClient($http, 'http://fit-chat:8081', 'secret', new NullLogger());

        $this->expectException(FitChatUnavailableException::class);
        $client->ask(new FitChatAskRequest(
            userId: 1,
            activityId: 104,
            message: 'Hi',
            storageBucket: 'racecoach-fits',
            featuresObjectKey: 'users/1/fits/x.features.json',
        ));
    }

    public function testAskMapsClientErrorToRejected(): void
    {
        $http = new MockHttpClient([
            new MockResponse('{"message":"message is required"}', ['http_code' => 400]),
        ]);
        $client = new FitChatClient($http, 'http://fit-chat:8081', 'secret', new NullLogger());

        try {
            $client->ask(new FitChatAskRequest(
                userId: 1,
                activityId: 104,
                message: 'Hi',
                storageBucket: 'racecoach-fits',
                featuresObjectKey: 'users/1/fits/x.features.json',
            ));
            self::fail('Expected FitChatRejectedException');
        } catch (FitChatRejectedException $exception) {
            self::assertSame(400, $exception->statusCode);
            self::assertSame('message is required', $exception->getMessage());
        }
    }

    public function testAskMapsMissingReplyToProtocolError(): void
    {
        $http = new MockHttpClient([
            new MockResponse('{"ok":true}', ['http_code' => 200]),
        ]);
        $client = new FitChatClient($http, 'http://fit-chat:8081', 'secret', new NullLogger());

        $this->expectException(FitChatProtocolException::class);
        $client->ask($this->sampleAsk());
    }

    public function testAskSendsBearerAndArtifactKeys(): void
    {
        $http = new MockHttpClient(function (string $method, string $url, array $options): MockResponse {
            self::assertSame('POST', $method);
            self::assertSame('http://fit-chat:8081/v1/chat', $url);

            $authorization = $options['normalized_headers']['authorization'][0]
                ?? $options['headers']['Authorization']
                ?? null;
            self::assertSame('Authorization: Bearer secret', $authorization);

            $body = json_decode($options['body'], true, 512, JSON_THROW_ON_ERROR);
            self::assertSame([
                'userId' => 7,
                'activityId' => 104,
                'message' => 'Tell me more',
                'storageBucket' => 'racecoach-fits',
                'featuresObjectKey' => 'users/7/a.features.json',
                'summaryObjectKey' => 'users/7/a.summary.json',
                'summary' => 'Headline note',
            ], $body);

            return new MockResponse(json_encode(['reply' => 'ok'], JSON_THROW_ON_ERROR));
        });

        $client = new FitChatClient($http, 'http://fit-chat:8081/', 'secret', new NullLogger());
        $client->ask(new FitChatAskRequest(
            userId: 7,
            activityId: 104,
            message: 'Tell me more',
            storageBucket: 'racecoach-fits',
            featuresObjectKey: 'users/7/a.features.json',
            summaryObjectKey: 'users/7/a.summary.json',
            summary: 'Headline note',
        ));
    }
}
