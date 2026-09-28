<?php

namespace App\Service\Storage;

use Aws\S3\Exception\S3Exception;
use Aws\S3\S3Client;
use RuntimeException;

/**
 * Thin wrapper around the AWS S3 client.
 */
final class ObjectStorage
{
    private readonly S3Client $client;

    public function __construct(
        private readonly string $endpoint,
        private readonly string $region,
        private readonly string $accessKey,
        private readonly string $secretKey,
        private readonly string $bucket,
        private readonly bool $usePathStyle,
    ) {
        $this->client = new S3Client([
            'version' => 'latest',
            'region' => $this->region,
            'endpoint' => $this->endpoint,
            'use_path_style_endpoint' => $this->usePathStyle,
            'credentials' => [
                'key' => $this->accessKey,
                'secret' => $this->secretKey,
            ],
        ]);
    }

    public function getBucket(): string
    {
        return $this->bucket;
    }

    /**
     * Uploads a local tempfile to object storage and returns the object key.
     */
    public function putFile(string $objectKey, string $sourcePath, string $contentType = 'application/octet-stream'): string
    {
        try {
            $this->client->putObject([
                'Bucket' => $this->bucket,
                'Key' => $objectKey,
                'SourceFile' => $sourcePath,
                'ContentType' => $contentType,
            ]);
        } catch (S3Exception $exception) {
            throw new RuntimeException(
                sprintf('Failed to upload object "%s" to bucket "%s": %s', $objectKey, $this->bucket, $exception->getAwsErrorMessage() ?: $exception->getMessage()),
                previous: $exception,
            );
        }

        return $objectKey;
    }
}
