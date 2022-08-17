# 测试认证 + 鉴权
1. user里面创建member
post  http://127.0.0.1:8050/keyauth/api/v1/token/issue
{
   "user_name": "member",
   "password": "123456"
}
2. 颁发token
post  http://127.0.0.1:8050/keyauth/api/v1/token/issue
{
   "user_name": "member",
   "password": "123456"
}
{
    "code": 0,
    "data": {
        "access_token": "ZIfaB uZyNsWDQUQyqD90LSW",
        "issue_at": 1659104548773,
        "update_at": 0,
        "update_by": "",
        "data": {
            " rante_type": "PASSWORD",
            "user_domain": "default",
            "user_name": "member",
            "password": ""
        },
        "access_token_expired_at": 1659105148773,
        "refresh_token": "fUeS Axavb Q 9HejTEhYHZB0tWeBrip",
        "refresh_token_expired_at": 1659107548773,
        "domian": "",
        "meta": null
    }
}
3. 查询secret
get http://127.0.0.1:8060/CMDB/api/v1/secret


# 测试权限接入后
post  http://127.0.0.1:8060/CMDB/api/v1/secret/
{
   "description": "权限测试",
   "allow_regions": ["*"],
   "api_key": "test-key",
   "api_secret": "test-key-secret"
}
{
    "code": 0,
    "data": {
        "id": "cbl4gsfj8ck4ga38h3g0",
        "create_at": 1659521137552,
        "data": {
            "description": "权限测试",
            "vendor": "ALIYUN",
            "allow_regions": [
                "*"
            ],
            "crendential_type": "API_KEY",
            "address": "",
            "api_key": "test-key",
            "api_secret": "@ciphered@9PBlOpFUF1C/JrxEj4JUVCkChNI1BTAMSnjd/ZaEYPA=",
            "request_rate": 5,
            "create_by": "member"
        }
    }
}