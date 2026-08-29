package com.example.sub

import android.content.Context
import android.content.SharedPreferences
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.json.JSONArray
import org.json.JSONObject
import java.net.HttpURLConnection
import java.net.URL

object ApiService {

    private const val BASE_URL = "http://127.0.0.1:8087/api"

    suspend fun register(context: Context, email: String, password: String, role: String = "user"): Boolean = withContext(Dispatchers.IO) {
        val url = URL("$BASE_URL/register")
        val conn = url.openConnection() as HttpURLConnection
        conn.requestMethod = "POST"
        conn.setRequestProperty("Content-Type", "application/json")
        conn.doOutput = true
        conn.connectTimeout = 5000
        conn.readTimeout = 5000

        val body = JSONObject().apply {
            put("email", email)
            put("password", password)
            put("role", role)
        }

        conn.outputStream.bufferedWriter().use { it.write(body.toString()) }

        return@withContext conn.responseCode == 201
    }
    suspend fun deleteTemplate(context: Context, id: Int): Boolean = withContext(Dispatchers.IO) {
        val url = URL("$BASE_URL/admin/templates/$id")
        val conn = url.openConnection() as HttpURLConnection
        conn.requestMethod = "DELETE"
        conn.setRequestProperty("Authorization", "Bearer ${TokenManager.getToken(context)}")
        conn.connectTimeout = 5000
        conn.readTimeout = 5000

        return@withContext conn.responseCode == 200
    }
    suspend fun updateTemplate(context: Context, id: Int, serviceName: String, price: Int): Boolean = withContext(Dispatchers.IO) {
        val url = URL("$BASE_URL/admin/templates/$id")
        val conn = url.openConnection() as HttpURLConnection
        conn.requestMethod = "PUT"
        conn.setRequestProperty("Content-Type", "application/json")
        conn.setRequestProperty("Authorization", "Bearer ${TokenManager.getToken(context)}")
        conn.doOutput = true
        conn.connectTimeout = 5000
        conn.readTimeout = 5000

        val body = JSONObject().apply {
            put("service_name", serviceName)
            put("price", price)
        }

        conn.outputStream.bufferedWriter().use { it.write(body.toString()) }

        return@withContext conn.responseCode == 200
    }
    suspend fun createTemplate(context: Context, serviceName: String, price: Int): Int? = withContext(Dispatchers.IO) {
        val url = URL("$BASE_URL/admin/templates")
        val conn = url.openConnection() as HttpURLConnection
        conn.requestMethod = "POST"
        conn.setRequestProperty("Content-Type", "application/json")
        conn.setRequestProperty("Authorization", "Bearer ${TokenManager.getToken(context)}")
        conn.doOutput = true
        conn.connectTimeout = 5000
        conn.readTimeout = 5000

        val body = JSONObject().apply {
            put("service_name", serviceName)
            put("price", price)
        }

        conn.outputStream.bufferedWriter().use { it.write(body.toString()) }

        return@withContext if (conn.responseCode == 201) {
            val json = conn.inputStream.bufferedReader().use { it.readText() }
            JSONObject(json).getInt("id")
        } else null
    }
    suspend fun getTemplates(context: Context): List<Template> = withContext(Dispatchers.IO) {
        val url = URL("$BASE_URL/templates")
        val conn = url.openConnection() as HttpURLConnection
        conn.requestMethod = "GET"
        conn.setRequestProperty("Authorization", "Bearer ${TokenManager.getToken(context)}")
        conn.connectTimeout = 5000
        conn.readTimeout = 5000

        return@withContext if (conn.responseCode == 200) {
            val json = conn.inputStream.bufferedReader().use { it.readText() }
            parseTemplates(json)
        } else {
            emptyList()
        }
    }
    suspend fun login(context: Context, email: String, password: String): Boolean =
        withContext(Dispatchers.IO) {
            val url = URL("$BASE_URL/login")
            val conn = url.openConnection() as HttpURLConnection
            conn.requestMethod = "POST"
            conn.setRequestProperty("Content-Type", "application/json")
            conn.doOutput = true
            conn.connectTimeout = 5000
            conn.readTimeout = 5000

            val body = JSONObject().apply {
                put("email", email)
                put("password", password)
            }

            conn.outputStream.bufferedWriter().use { it.write(body.toString()) }

            return@withContext if (conn.responseCode == 200) {
                val json = conn.inputStream.bufferedReader().use { it.readText() }
                val token = JSONObject(json).getString("token")
                TokenManager.saveToken(context, token)
                true
            } else false
        }

    suspend fun getSubscriptions(context: Context): List<Subscription> = withContext(Dispatchers.IO) {
        val url = URL("$BASE_URL/subscriptions")
        val conn = url.openConnection() as HttpURLConnection
        conn.requestMethod = "GET"
        conn.setRequestProperty("Authorization", "Bearer ${TokenManager.getToken(context)}")
        conn.connectTimeout = 5000
        conn.readTimeout = 5000

        return@withContext if (conn.responseCode == 200) {
            val json = conn.inputStream.bufferedReader().use { it.readText() }
            parseSubscriptions(json)
        } else {
            emptyList()
        }
    }

    suspend fun createSubscriptionFromTemplate(context: Context, templateId: Int, startDate: String, endDate: String): Int? = withContext(Dispatchers.IO) {
        val url = URL("$BASE_URL/subscriptions")
        val conn = url.openConnection() as HttpURLConnection
        conn.requestMethod = "POST"
        conn.setRequestProperty("Content-Type", "application/json")
        conn.setRequestProperty("Authorization", "Bearer ${TokenManager.getToken(context)}")
        conn.doOutput = true
        conn.connectTimeout = 5000
        conn.readTimeout = 5000

        val body = JSONObject().apply {
            put("template_id", templateId)
            put("start_date", startDate)
            put("end_date", endDate)
        }

        conn.outputStream.bufferedWriter().use { it.write(body.toString()) }

        return@withContext if (conn.responseCode == 201) {
            val json = conn.inputStream.bufferedReader().use { it.readText() }
            JSONObject(json).getInt("id")
        } else null
    }
    suspend fun createSubscription(context: Context, sub: Subscription): Int? = withContext(Dispatchers.IO) {
        val url = URL("$BASE_URL/subscriptions")
        val conn = url.openConnection() as HttpURLConnection
        conn.requestMethod = "POST"
        conn.setRequestProperty("Content-Type", "application/json")
        conn.setRequestProperty("Authorization", "Bearer ${TokenManager.getToken(context)}")
        conn.doOutput = true
        conn.connectTimeout = 5000
        conn.readTimeout = 5000

        val body = JSONObject().apply {
            put("service_name", sub.serviceName)
            put("price", sub.price)
            put("user_id", sub.userId)
            put("start_date", sub.startDate)
            put("end_date", sub.endDate)
        }

        conn.outputStream.bufferedWriter().use { it.write(body.toString()) }

        return@withContext if (conn.responseCode == 201) {
            val json = conn.inputStream.bufferedReader().use { it.readText() }
            JSONObject(json).getInt("id")
        } else null
    }

    suspend fun deleteSubscription(context: Context, id: Int): Boolean = withContext(Dispatchers.IO) {
        val url = URL("$BASE_URL/subscriptions/$id")
        val conn = url.openConnection() as HttpURLConnection
        conn.requestMethod = "DELETE"
        conn.setRequestProperty("Authorization", "Bearer ${TokenManager.getToken(context)}")
        conn.connectTimeout = 5000
        conn.readTimeout = 5000

        return@withContext conn.responseCode == 200
    }

    suspend fun getTotalCost(context: Context, startDate: String, endDate: String, userId: String = ""): Int? =
        withContext(Dispatchers.IO) {
            var urlString = "$BASE_URL/subscriptions/total-cost?start_date=$startDate&end_date=$endDate"
            if (userId.isNotBlank()) urlString += "&user_id=$userId"

            val url = URL(urlString)
            val conn = url.openConnection() as HttpURLConnection
            conn.requestMethod = "GET"
            conn.setRequestProperty("Authorization", "Bearer ${TokenManager.getToken(context)}")
            conn.connectTimeout = 5000
            conn.readTimeout = 5000

            return@withContext if (conn.responseCode == 200) {
                val json = conn.inputStream.bufferedReader().use { it.readText() }
                JSONObject(json).getInt("total")
            } else null
        }

    suspend fun getSubscription(context: Context, id: Int): Subscription? = withContext(Dispatchers.IO) {
        val url = URL("$BASE_URL/subscriptions/$id")
        val conn = url.openConnection() as HttpURLConnection
        conn.requestMethod = "GET"
        conn.setRequestProperty("Authorization", "Bearer ${TokenManager.getToken(context)}")
        conn.connectTimeout = 5000
        conn.readTimeout = 5000

        return@withContext if (conn.responseCode == 200) {
            val json = conn.inputStream.bufferedReader().use { it.readText() }
            val obj = JSONObject(json)
            Subscription(
                id = obj.getInt("id"),
                serviceName = obj.getString("service_name"),
                price = obj.getInt("price"),
                userId = obj.getString("user_id"),
                startDate = obj.getString("start_date"),
                endDate = obj.optString("end_date", "")
            )
        } else null
    }
suspend fun updateSubscriptionFull(context: Context, id: Int, serviceName: String, price: Int, userId: String, startDate: String, endDate: String): Boolean = withContext(Dispatchers.IO) {
    val url = URL("$BASE_URL/subscriptions/$id")
    val conn = url.openConnection() as HttpURLConnection
    conn.requestMethod = "PUT"
    conn.setRequestProperty("Content-Type", "application/json")
    conn.setRequestProperty("Authorization", "Bearer ${TokenManager.getToken(context)}")
    conn.doOutput = true
    conn.connectTimeout = 5000
    conn.readTimeout = 5000

    val body = JSONObject().apply {
        put("service_name", serviceName)
        put("price", price)
        put("user_id", userId)
        put("start_date", startDate)
        put("end_date", endDate)
    }

    conn.outputStream.bufferedWriter().use { it.write(body.toString()) }

    return@withContext conn.responseCode == 200
}
    suspend fun updateSubscription(context: Context, id: Int, startDate: String, endDate: String): Boolean =
        withContext(Dispatchers.IO) {
            val url = URL("$BASE_URL/subscriptions/$id")
            val conn = url.openConnection() as HttpURLConnection
            conn.requestMethod = "PUT"
            conn.setRequestProperty("Content-Type", "application/json")
            conn.setRequestProperty("Authorization", "Bearer ${TokenManager.getToken(context)}")
            conn.doOutput = true
            conn.connectTimeout = 5000
            conn.readTimeout = 5000

            val body = JSONObject().apply {
                if (startDate.isNotEmpty()) put("start_date", startDate)
                if (endDate.isNotEmpty()) put("end_date", endDate)
            }

            conn.outputStream.bufferedWriter().use { it.write(body.toString()) }

            return@withContext conn.responseCode == 200
        }

    private fun parseTemplates(json: String): List<Template> {
    if (json == "null" || json.isEmpty()) return emptyList()
        val result = mutableListOf<Template>()
        val arr = JSONArray(json)
        for (i in 0 until arr.length()) {
            val obj = arr.getJSONObject(i)
            result.add(
                Template(
                    id = obj.getInt("id"),
                    serviceName = obj.getString("service_name"),
                    price = obj.getInt("price")
                )
            )
        }
        return result
    }
    private fun parseSubscriptions(json: String): List<Subscription> {
        val result = mutableListOf<Subscription>()
        val arr = JSONArray(json)
        for (i in 0 until arr.length()) {
            val obj = arr.getJSONObject(i)
            result.add(
                Subscription(
                    id = obj.getInt("id"),
                    serviceName = obj.getString("service_name"),
                    price = obj.getInt("price"),
                    userId = obj.getString("user_id"),
                    startDate = obj.getString("start_date"),
                    endDate = obj.optString("end_date", "")
                )
            )
        }
        return result
    }
}

object TokenManager {
    private const val PREF_NAME = "auth"
    private const val TOKEN_KEY = "jwt_token"

    fun saveToken(context: Context, token: String) {
        val prefs: SharedPreferences = context.getSharedPreferences(PREF_NAME, Context.MODE_PRIVATE)
        prefs.edit().putString(TOKEN_KEY, token).apply()
    }

    fun getToken(context: Context): String {
        val prefs: SharedPreferences = context.getSharedPreferences(PREF_NAME, Context.MODE_PRIVATE)
        return prefs.getString(TOKEN_KEY, "") ?: ""
    }
}
