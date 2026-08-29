package com.example.sub

import android.os.Bundle
import android.widget.Button
import android.widget.EditText
import android.widget.TextView
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.text.SimpleDateFormat
import java.util.*

class EditSubscriptionActivity : AppCompatActivity() {

    private lateinit var tvServiceName: TextView
    private lateinit var tvPrice: TextView
    private lateinit var etStartDate: EditText
    private lateinit var etEndDate: EditText
    private lateinit var btnSave: Button
    private lateinit var btnCancel: Button

    private val scope = CoroutineScope(Dispatchers.Main)
    private var subscriptionId: Int = 0
    private var originalStartDate: String = ""
    private var originalServiceName: String = ""
    private var originalPrice: Int = 0
    private var originalUserId: String = ""

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_edit_subscription)

        tvServiceName = findViewById(R.id.tvServiceName)
        tvPrice = findViewById(R.id.tvPrice)
        etStartDate = findViewById(R.id.etStartDate)
        etEndDate = findViewById(R.id.etEndDate)
        btnSave = findViewById(R.id.btnSave)
        btnCancel = findViewById(R.id.btnCancel)

        subscriptionId = intent.getIntExtra("subscription_id", 0)
        if (subscriptionId == 0) {
            Toast.makeText(this, "Ошибка: ID не передан", Toast.LENGTH_SHORT).show()
            finish()
            return
        }

        loadSubscription()

        btnSave.setOnClickListener {
            val startDate = etStartDate.text.toString().trim()
            val endDate = etEndDate.text.toString().trim()

            if (startDate.isEmpty()) {
                Toast.makeText(this, "Введите дату начала", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }

            scope.launch {
                val success = ApiService.updateSubscriptionFull(
                    this@EditSubscriptionActivity,
                    subscriptionId,
                    originalServiceName,
                    originalPrice,
                    originalUserId,
                    startDate,
                    endDate
                )
                withContext(Dispatchers.Main) {
                    if (success) {
                        Toast.makeText(this@EditSubscriptionActivity, "Подписка обновлена!", Toast.LENGTH_LONG).show()
                        finish()
                    } else {
                        Toast.makeText(this@EditSubscriptionActivity, "Ошибка обновления", Toast.LENGTH_SHORT).show()
                    }
                }
            }
        }

        btnCancel.setOnClickListener {
            finish()
        }
    }

    private fun loadSubscription() {
        scope.launch {
            val sub = ApiService.getSubscription(this@EditSubscriptionActivity, subscriptionId)
            withContext(Dispatchers.Main) {
                if (sub != null) {
                    tvServiceName.text = "📌 ${sub.serviceName}"
                    tvPrice.text = "💰 ${sub.price} ₽"
                    etStartDate.setText(sub.startDate)
                    etEndDate.setText(sub.endDate)
                    originalStartDate = sub.startDate
                    originalServiceName = sub.serviceName
                    originalPrice = sub.price
                    originalUserId = sub.userId

                    // Блокируем start_date, если она уже наступила
                    if (isStartDateInPast(sub.startDate)) {
                        etStartDate.isEnabled = false
                        etStartDate.alpha = 0.5f
                        Toast.makeText(
                            this@EditSubscriptionActivity,
                            "Дата начала уже наступила, изменение запрещено",
                            Toast.LENGTH_LONG
                        ).show()
                    }
                } else {
                    Toast.makeText(this@EditSubscriptionActivity, "Подписка не найдена", Toast.LENGTH_SHORT).show()
                    finish()
                }
            }
        }
    }

    private fun isStartDateInPast(dateStr: String): Boolean {
        return try {
            val sdf = SimpleDateFormat("MM-yyyy", Locale.getDefault())
            val date = sdf.parse(dateStr) ?: return false
            val today = Calendar.getInstance()
            today.set(Calendar.DAY_OF_MONTH, 1)
            today.set(Calendar.HOUR_OF_DAY, 0)
            today.set(Calendar.MINUTE, 0)
            today.set(Calendar.SECOND, 0)
            today.set(Calendar.MILLISECOND, 0)
            date.before(today.time)
        } catch (e: Exception) {
            false
        }
    }
}
