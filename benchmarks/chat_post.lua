-- wrk script for POST /v1/chat/completions
wrk.method = "POST"
wrk.body   = '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"Hi"}]}'
wrk.headers["Content-Type"] = "application/json"
wrk.headers["Authorization"] = "Bearer test-key"

-- 使用随机化的简单消息以避免缓存
request = function()
 local messages = {
   '{"role":"user","content":"Hello"}',
   '{"role":"user","content":"Hi there"}',
   '{"role":"user","content":"How are you?"}',
   '{"role":"user","content":"Test message"}',
   '{"role":"user","content":"Greetings"}'
 }
 local msg = messages[math.random(#messages)]
 local body = string.format('{"model":"gpt-3.5-turbo","messages":[%s]}', msg)
 return wrk.format(nil, nil, nil, body)
end
