@echo off
REM 算能云 OpenAI 格式测试脚本

REM 配置
set API_BASE=http://localhost:3000
set API_KEY=your-api-key-here

echo ==========================================
echo 算能云 OpenAI 格式兼容性测试
echo ==========================================
echo.

REM 测试 1: OpenAI 格式图片生成
echo 测试 1: OpenAI 格式图片生成 (Qwen-Image-Plus)
echo ------------------------------------------
curl -X POST "%API_BASE%/v1/images/generations" ^
  -H "Content-Type: application/json" ^
  -H "Authorization: Bearer %API_KEY%" ^
  -d "{\"model\": \"Qwen-Image-Plus\", \"prompt\": \"一只可爱的猫咪在草地上玩耍\", \"size\": \"1024x1024\", \"watermark\": false}"
echo.
echo.

REM 测试 2: OpenAI 格式图片生成（带扩展参数）
echo 测试 2: OpenAI 格式图片生成（带扩展参数）
echo ------------------------------------------
curl -X POST "%API_BASE%/v1/images/generations" ^
  -H "Content-Type: application/json" ^
  -H "Authorization: Bearer %API_KEY%" ^
  -d "{\"model\": \"Qwen-Image-Plus\", \"prompt\": \"一只可爱的猫咪在草地上玩耍\", \"size\": \"1024x1024\", \"watermark\": false, \"negative_prompt\": \"模糊，低质量\", \"seed\": 12345, \"prompt_extend\": true}"
echo.
echo.

REM 测试 3: OpenAI 格式视频生成 (T2V)
echo 测试 3: OpenAI 格式视频生成 (Wan2.6-T2V)
echo ------------------------------------------
curl -X POST "%API_BASE%/v1/videos" ^
  -H "Content-Type: application/json" ^
  -H "Authorization: Bearer %API_KEY%" ^
  -d "{\"model\": \"Wan2.6-T2V\", \"prompt\": \"一只猫咪在草地上奔跑\", \"duration\": 5.0, \"width\": 1280, \"height\": 720, \"seed\": 12345}"
echo.
echo.

REM 测试 4: OpenAI 格式视频生成（带扩展参数）
echo 测试 4: OpenAI 格式视频生成（带扩展参数）
echo ------------------------------------------
curl -X POST "%API_BASE%/v1/videos" ^
  -H "Content-Type: application/json" ^
  -H "Authorization: Bearer %API_KEY%" ^
  -d "{\"model\": \"Wan2.6-T2V\", \"prompt\": \"一只猫咪在草地上奔跑\", \"duration\": 5.0, \"width\": 1280, \"height\": 720, \"seed\": 12345, \"metadata\": {\"negative_prompt\": \"模糊，抖动\", \"subdivision_level\": \"high\", \"file_format\": \"mp4\"}}"
echo.
echo.

REM 测试 5: OpenAI 格式图生视频 (I2V)
echo 测试 5: OpenAI 格式图生视频 (Wan2.6-I2V)
echo ------------------------------------------
curl -X POST "%API_BASE%/v1/videos" ^
  -H "Content-Type: application/json" ^
  -H "Authorization: Bearer %API_KEY%" ^
  -d "{\"model\": \"Wan2.6-I2V\", \"prompt\": \"猫咪开始奔跑\", \"image\": \"https://example.com/cat.jpg\", \"duration\": 5.0}"
echo.
echo.

REM 测试 6: 算能原生格式图片生成（验证兼容性）
echo 测试 6: 算能原生格式图片生成（验证兼容性）
echo ------------------------------------------
curl -X POST "%API_BASE%/imagegenerator/task" ^
  -H "Content-Type: application/json" ^
  -H "Authorization: Bearer %API_KEY%" ^
  -d "{\"model\": \"Qwen-Image-Plus\", \"input\": {\"prompt\": \"一只可爱的猫咪在草地上玩耍\", \"negative_prompt\": \"模糊，低质量\"}, \"parameters\": {\"size\": \"1024x1024\", \"seed\": 12345, \"watermark\": false}}"
echo.
echo.

REM 测试 7: 算能原生格式视频生成（验证兼容性）
echo 测试 7: 算能原生格式视频生成（验证兼容性）
echo ------------------------------------------
curl -X POST "%API_BASE%/videogenerator/generate" ^
  -H "Content-Type: application/json" ^
  -H "Authorization: Bearer %API_KEY%" ^
  -d "{\"model\": \"Wan2.6-T2V\", \"content\": [{\"type\": \"text\", \"text\": \"一只猫咪在草地上奔跑\", \"negative_prompt\": \"模糊，抖动\"}], \"parameters\": {\"size\": \"1280x720\", \"duration\": 5, \"seed\": \"12345\"}}"
echo.
echo.

echo ==========================================
echo 测试完成
echo ==========================================
pause
