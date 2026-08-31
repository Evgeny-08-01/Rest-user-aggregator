package com.example.sub

import android.os.Bundle
import android.widget.*
import androidx.appcompat.app.AppCompatActivity
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import org.json.JSONObject

class AddSubscriptionActivity : AppCompatActivity() {

    private lateinit var spinnerTemplates: Spinner
    private lateinit var etStartDate: EditText
    private lateinit var etEndDate: EditText
    private lateinit var btnSave: Button
    private lateinit var btnCancel: Button

    private val scope = CoroutineScope(Dispatchers.Main)
    private var templates: List<Template> = emptyList()
    private var selectedTemplateId: Int = -1

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_add_subscription)

        spinnerTemplates = findViewById(R.id.spinnerTemplates)
        etStartDate = findViewById(R.id.etStartDate)
        etEndDate = findViewById(R.id.etEndDate)
        btnSave = findViewById(R.id.btnSave)
        btnCancel = findViewById(R.id.btnCancel)

        loadTemplates()

        btnSave.setOnClickListener {
            val startDate = etStartDate.text.toString().trim()
            val endDate = etEndDate.text.toString().trim()

            if (selectedTemplateId == -1) {
                Toast.makeText(this, "Выберите шаблон", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            if (startDate.isEmpty()) {
                Toast.makeText(this, "Введите дату начала", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }

            btnSave.isEnabled = false

            scope.launch {
                val result = ApiService.createSubscriptionFromTemplate(
                    this@AddSubscriptionActivity,
                    selectedTemplateId,
                    startDate,
                    endDate
                )
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

    private fun loadTemplates() {
        scope.launch {
            templates = ApiService.getTemplates(this@AddSubscriptionActivity)
            withContext(Dispatchers.Main) {
                if (templates.isEmpty()) {
                    Toast.makeText(this@AddSubscriptionActivity, "Нет доступных шаблонов", Toast.LENGTH_SHORT).show()
                    finish()
                    return@withContext
                }

                val adapter = ArrayAdapter(
                    this@AddSubscriptionActivity,
                    android.R.layout.simple_spinner_item,
                    templates.map { "${it.serviceName} (${it.price} ₽)" }
                )
                adapter.setDropDownViewResource(android.R.layout.simple_spinner_dropdown_item)
                spinnerTemplates.adapter = adapter

        android.util.Log.d("CreateSubscription", "Selected template ID: $selectedTemplateId")
                spinnerTemplates.onItemSelectedListener = object : AdapterView.OnItemSelectedListener {
                    override fun onItemSelected(parent: AdapterView<*>?, view: android.view.View?, position: Int, id: Long) {
                        selectedTemplateId = templates[position].id
                    }
                    override fun onNothingSelected(parent: AdapterView<*>?) {
                        selectedTemplateId = -1
                    }
                }
            }
        }
    }
}
