package com.example.sub

import android.content.Context
import org.json.JSONObject

data class User(val email: String, val role: String)

fun getUser(context: Context): User? {
    val token = TokenManager.getToken(context)
    if (token.isBlank()) return null
    return try {
        val parts = token.split(".")
        if (parts.size != 3) return null
        val payload = String(android.util.Base64.decode(parts[1], android.util.Base64.URL_SAFE))
        val json = JSONObject(payload)
        User(
            email = json.optString("email", ""),
            role = json.optString("role", "")
        )
    } catch (e: Exception) {
        null
    }
}
