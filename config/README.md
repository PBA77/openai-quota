# Configuration Directory

This directory contains configuration files for the OpenAI Quota Proxy.

## Files

### `model_pricing.csv`
CSV file containing pricing information for OpenAI models.

**Format:**
```csv
model,version,input,cached_input,output
gpt-4o,gpt-4o-2024-08-06,2.5,1.25,10.0
gpt-4o-mini,gpt-4o-mini-2024-07-18,0.15,0.075,0.6
```

**Columns:**
- `model`: Model name (e.g., "gpt-4o")
- `version`: Specific model version 
- `input`: Input token price per 1M tokens (USD)
- `cached_input`: Cached input token price per 1M tokens (USD)
- `output`: Output token price per 1M tokens (USD)

**Usage:**
The application loads this file at startup using the `-pricing` flag:
```bash
./openai-quota -pricing=config/model_pricing.csv
```

### `app.env`
Environment configuration template with default settings.

**Contains:**
- Default port and quota settings
- Production vs development configurations
- Logging preferences
- Security settings
- Performance tuning parameters

**Note:** This is a template file. Copy to `.env` for actual configuration.

## Adding New Models

To add a new model to the pricing configuration:

1. Open `model_pricing.csv`
2. Add a new line with the model information
3. Restart the application to load new pricing

**Example:**
```csv
new-model,new-model-v1,1.5,0.75,5.0
```

## Configuration Priority

1. Command-line flags (highest priority)
2. Environment variables
3. Configuration files
4. Default values (lowest priority)

## Security Notes

- Keep pricing files up to date with OpenAI's current rates
- Validate all pricing data before deployment
- Use appropriate file permissions for configuration files
- Consider encrypting sensitive configuration in production
