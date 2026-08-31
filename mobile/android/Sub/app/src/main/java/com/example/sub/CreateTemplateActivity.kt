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

class CreateTemplateActivity : AppCompatActivity() {

    private lateinit var etName: EditText
    private lateinit var etPrice: EditText
    private lateinit var btnSave: Button
    private lateinit var btnCancel: Button

    private val scope = CoroutineScope(Dispatchers.Main)
    private var templateId: Int = 0

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_create_template)

        etName = findViewById(R.id.etTemplateName)
        etPrice = findViewById(R.id.etTemplatePrice)
        btnSave = findViewById(R.id.btnSaveTemplate)
        btnCancel = findViewById(R.id.btnCancelTemplate)

        // Проверяем, редактирование или создание
        templateId = intent.getIntExtra("template_id", 0)
        if (templateId > 0) {
            loadTemplate()
        }

        btnSave.setOnClickListener {
            val name = etName.text.toString().trim()
            val priceStr = etPrice.text.toString().trim()

            if (name.isEmpty() || priceStr.isEmpty()) {
                Toast.makeText(this, "Заполните все поля", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }

            val price = priceStr.toIntOrNull()
            if (price == null || price < 0) {
                Toast.makeText(this, "Цена должна быть >= 0", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }

            btnSave.isEnabled = false

            scope.launch {
                val success = if (templateId > 0) {
                    ApiService.updateTemplate(this@CreateTemplateActivity, templateId, name, price)
                } else {
                    ApiService.createTemplate(this@CreateTemplateActivity, name, price) != null
                }
                withContext(Dispatchers.Main) {
                    btnSave.isEnabled = true
                    if (success) {
                        Toast.makeText(
                            this@CreateTemplateActivity,
                            if (templateId > 0) "Шаблон обновлён" else "Шаблон создан",
                            Toast.LENGTH_LONG
                        ).show()
                        finish()
                    } else {
                        Toast.makeText(this@CreateTemplateActivity, "Ошибка сохранения", Toast.LENGTH_SHORT).show()
                    }
                }
            }
        }

        btnCancel.setOnClickListener {
            finish()
        }
    }

    private fun loadTemplate() {
        scope.launch {
            val templates = ApiService.getTemplates(this@CreateTemplateActivity)
            val template = templates.find { it.id == templateId }
            withContext(Dispatchers.Main) {
                if (template != null) {
                    etName.setText(template.serviceName)
                    etPrice.setText(template.price.toString())
                } else {
                    Toast.makeText(this@CreateTemplateActivity, "Шаблон не найден", Toast.LENGTH_SHORT).show()
                    finish()
                }
            }
        }
    }
}
