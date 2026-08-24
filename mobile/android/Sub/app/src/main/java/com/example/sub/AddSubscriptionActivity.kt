package com.example.sub

import android.os.Bundle
import android.widget.Button
import android.widget.EditText
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

class AddSubscriptionActivity : AppCompatActivity() {

    private lateinit var etServiceName: EditText
    private lateinit var etPrice: EditText
    private lateinit var etStartDate: EditText
    private lateinit var etEndDate: EditText
    private lateinit var btnSave: Button
    private lateinit var btnCancel: Button

    private val scope = CoroutineScope(Dispatchers.Main)

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_add_subscription)

        etServiceName = findViewById(R.id.etServiceName)
        etPrice = findViewById(R.id.etPrice)
        etStartDate = findViewById(R.id.etStartDate)
        etEndDate = findViewById(R.id.etEndDate)
        btnSave = findViewById(R.id.btnSave)
        btnCancel = findViewById(R.id.btnCancel)

        // Значения по умолчанию для быстрого тестирования
        etServiceName.setText("Авито")
        etPrice.setText("300")
        etStartDate.setText("08-2025")
        etEndDate.setText("12-2025")

        btnSave.setOnClickListener {
            val serviceName = etServiceName.text.toString().trim()
            val priceStr = etPrice.text.toString().trim()
            val startDate = etStartDate.text.toString().trim()
            val endDate = etEndDate.text.toString().trim()

            if (serviceName.isEmpty() || priceStr.isEmpty() || startDate.isEmpty()) {
                Toast.makeText(this, "Заполните все обязательные поля", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }

            val price = priceStr.toIntOrNull()
            if (price == null || price < 0) {
                Toast.makeText(this, "Цена должна быть >= 0", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }

            // Берём user_id из токена
            val token = TokenManager.getToken(this)
            val userId = extractUserIdFromToken(token)

            if (userId.isNullOrEmpty()) {
                Toast.makeText(this, "Ошибка: не удалось определить user_id", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }

            val sub = Subscription(
                id = 0,
                serviceName = serviceName,
                price = price,
                userId = userId,
                startDate = startDate,
                endDate = endDate
            )

            btnSave.isEnabled = false

            scope.launch {
                val result = ApiService.createSubscription(this@AddSubscriptionActivity, sub)
                withContext(Dispatchers.Main) {
                    btnSave.isEnabled = true
                    if (result != null) {
                        Toast.makeText(this@AddSubscriptionActivity, "Подписка создана! ID: $result", Toast.LENGTH_LONG).show()
                        finish()
                    } else {
                        Toast.makeText(this@AddSubscriptionActivity, "Ошибка создания", Toast.LENGTH_SHORT).show()
                    }
                }
            }
        }

        btnCancel.setOnClickListener {
            finish()
        }
    }

    // Парсинг user_id из JWT-токена (без проверки подписи)
    private fun extractUserIdFromToken(token: String): String? {
        return try {
            val parts = token.split(".")
            if (parts.size != 3) return null
            val payload = String(android.util.Base64.decode(parts[1], android.util.Base64.URL_SAFE))
            val json = org.json.JSONObject(payload)
            json.optString("user_id", "")
        } catch (e: Exception) {
            null
        }
    }
}
