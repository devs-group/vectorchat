import os
import tempfile
from pathlib import Path

from fastapi import FastAPI, File, HTTPException, UploadFile
from fastapi.responses import JSONResponse, PlainTextResponse
from markitdown import MarkItDown

app = FastAPI(title="MarkItDown Service")
converter = MarkItDown()


@app.post("/convert", response_class=PlainTextResponse)
async def convert_file(file: UploadFile = File(...)) -> str:
    contents = await file.read()
    if not contents:
        raise HTTPException(status_code=400, detail="Uploaded file is empty.")

    suffix = Path(file.filename or "").suffix
    with tempfile.NamedTemporaryFile(delete=False, suffix=suffix) as tmp_file:
        tmp_file.write(contents)
        tmp_path = tmp_file.name

    try:
        converted = converter.convert(tmp_path)
    except Exception as exc:  # pragma: no cover - third-party errors are reported
        print(f"Conversion error: {exc}")
        raise HTTPException(status_code=422, detail=f"Conversion failed: {exc}") from exc
    finally:
        try:
            os.unlink(tmp_path)
        except OSError:
            pass

    markdown = _to_markdown_text(converted)
    if markdown is None:
        raise HTTPException(status_code=500, detail="Conversion returned an unexpected format.")

    return markdown


@app.get("/supported-extensions")
def supported_extensions() -> JSONResponse:
    extensions = _supported_extensions()
    return JSONResponse({"extensions": extensions})


def _to_markdown_text(converted) -> str | None:
    if converted is None:
        return None

    if isinstance(converted, str):
        return converted

    text_attr = getattr(converted, "text", None)
    if isinstance(text_attr, str):
        return text_attr

    markdown_attr = getattr(converted, "markdown", None)
    if isinstance(markdown_attr, str):
        return markdown_attr

    if isinstance(converted, (tuple, list)) and converted:
        last = converted[-1]
        if isinstance(last, str):
            return last

    return str(converted)


def _supported_extensions() -> list[str]:
    # Try _converters attribute first
    converters = getattr(converter, "_converters", None)

    extensions = set()
    if converters and isinstance(converters, list):
        # Each item in _converters is a ConverterRegistration with a .converter attribute
        for conv_reg in converters:
            if hasattr(conv_reg, 'converter'):
                actual_converter = conv_reg.converter
                converter_name = type(actual_converter).__name__.lower()

                # Map converter class names to extensions
                if 'pdf' in converter_name:
                    extensions.add('.pdf')
                elif 'docx' in converter_name:
                    extensions.add('.docx')
                elif 'doc' in converter_name and 'docx' not in converter_name:
                    extensions.add('.doc')
                elif 'pptx' in converter_name:
                    extensions.add('.pptx')
                elif 'ppt' in converter_name and 'pptx' not in converter_name:
                    extensions.add('.ppt')
                elif 'xlsx' in converter_name:
                    extensions.add('.xlsx')
                elif 'xls' in converter_name and 'xlsx' not in converter_name:
                    extensions.add('.xls')
                elif 'csv' in converter_name:
                    extensions.add('.csv')
                elif 'html' in converter_name:
                    extensions.update(['.html', '.htm'])
                elif 'epub' in converter_name:
                    extensions.add('.epub')
                elif 'ipynb' in converter_name:
                    extensions.add('.ipynb')
                elif 'image' in converter_name:
                    extensions.update(['.png', '.jpg', '.jpeg', '.gif', '.bmp', '.tiff', '.webp'])
                elif 'audio' in converter_name:
                    extensions.update(['.mp3', '.wav', '.m4a', '.flac'])
                elif 'msg' in converter_name:
                    extensions.add('.msg')
                elif 'plaintext' in converter_name:
                    extensions.update(['.txt', '.md', '.rst', '.log'])
                elif 'zip' in converter_name:
                    extensions.add('.zip')
                elif 'rss' in converter_name:
                    extensions.update(['.xml', '.rss'])
                elif 'json' in converter_name:
                    extensions.add('.json')

    # Add some common extensions that might not be covered
    extensions.update(['.yaml', '.yml', '.tsv', '.rtf', '.eml', '.tex'])

    # If we still don't have extensions, use a fallback list
    if not extensions:
        extensions = {
            '.pdf', '.doc', '.docx', '.ppt', '.pptx', '.xls', '.xlsx',
            '.txt', '.md', '.html', '.htm', '.csv', '.tsv', '.json',
            '.xml', '.yaml', '.yml', '.ipynb', '.rst', '.tex', '.rtf',
            '.eml', '.msg', '.epub', '.zip'
        }

    return sorted(extensions)
