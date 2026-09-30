<?php

namespace App\Service\Fit;

use Symfony\Component\HttpFoundation\File\UploadedFile;
use Symfony\Component\HttpKernel\Exception\BadRequestHttpException;

/**
 * Validates uploaded FIT files before they leave the Symfony boundary.
 */
final class FitFileValidator
{
    private const FIT_SIGNATURE = '.FIT';
    private const MIN_HEADER_BYTES = 12;

    public function __construct(
        private readonly int $maxBytes,
    ) {
    }

    public function validate(UploadedFile $file): void
    {
        if (!$file->isValid()) {
            throw new BadRequestHttpException(sprintf('Upload failed: %s', $file->getErrorMessage() ?: 'unknown error'));
        }

        $originalName = $file->getClientOriginalName();
        if (!preg_match('/\.fit$/i', $originalName)) {
            throw new BadRequestHttpException('Only .fit files are accepted.');
        }

        // Reject double extensions like activity.php.fit
        $basename = pathinfo($originalName, PATHINFO_FILENAME);
        if ('' === $basename || 1 === preg_match('/[\\\\\\/\\x00]/', $basename)) {
            throw new BadRequestHttpException('Invalid .fit filename.');
        }

        $size = $file->getSize();
        if (false === $size || $size <= 0) {
            throw new BadRequestHttpException('The uploaded .fit file is empty.');
        }

        if ($size > $this->maxBytes) {
            throw new BadRequestHttpException(sprintf('The .fit file exceeds the maximum size of %d bytes.', $this->maxBytes));
        }

        $pathname = $file->getPathname();
        $handle = fopen($pathname, 'rb');
        if (false === $handle) {
            throw new BadRequestHttpException('Unable to read the uploaded .fit file.');
        }

        try {
            $header = fread($handle, self::MIN_HEADER_BYTES);
        } finally {
            fclose($handle);
        }

        if (false === $header || strlen($header) < self::MIN_HEADER_BYTES) {
            throw new BadRequestHttpException('The uploaded file is too small to be a valid .fit activity.');
        }

        $headerSize = ord($header[0]);
        if ($headerSize < self::MIN_HEADER_BYTES || $headerSize > 64) {
            throw new BadRequestHttpException('Invalid .fit header size.');
        }

        $signature = substr($header, 8, 4);
        if (self::FIT_SIGNATURE !== $signature) {
            throw new BadRequestHttpException('The uploaded file is not a valid .fit activity (missing .FIT signature).');
        }
    }
}
