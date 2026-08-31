package com.example.sub

import android.os.Bundle
import android.widget.Button
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.recyclerview.widget.LinearLayoutManager
import androidx.recyclerview.widget.RecyclerView
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

class TemplatesActivity : AppCompatActivity() {

    private lateinit var recyclerView: RecyclerView
    private lateinit var adapter: TemplateAdapter
    private lateinit var btnAdd: Button

    private val scope = CoroutineScope(Dispatchers.Main)

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_templates)

        recyclerView = findViewById(R.id.recyclerViewTemplates)
        btnAdd = findViewById(R.id.btnAddTemplate)

        adapter = TemplateAdapter(
            items = emptyList(),
            onEdit = { template ->
                // TODO: открыть форму редактирования
                startActivity(android.content.Intent(this, CreateTemplateActivity::class.java).putExtra("template_id", template.id))
            },
            onDelete = { template ->
                deleteTemplate(template.id)
            }
        )

        recyclerView.layoutManager = LinearLayoutManager(this)
        recyclerView.adapter = adapter

        btnAdd.setOnClickListener {
            // TODO: открыть форму создания шаблона
            startActivity(android.content.Intent(this, CreateTemplateActivity::class.java))
        }

        loadTemplates()
    }

    override fun onResume() {
        super.onResume()
        loadTemplates()
    }

    private fun loadTemplates() {
        scope.launch {
            val list = ApiService.getTemplates(this@TemplatesActivity)
            withContext(Dispatchers.Main) {
                adapter.updateData(list)
                if (list.isEmpty()) {
                    Toast.makeText(this@TemplatesActivity, "Нет шаблонов", Toast.LENGTH_SHORT).show()
                }
            }
        }
    }

    private fun deleteTemplate(id: Int) {
        scope.launch {
            val success = ApiService.deleteTemplate(this@TemplatesActivity, id)
            withContext(Dispatchers.Main) {
                if (success) {
                    Toast.makeText(this@TemplatesActivity, "Шаблон удалён", Toast.LENGTH_SHORT).show()
                    loadTemplates()
                } else {
                    Toast.makeText(this@TemplatesActivity, "Ошибка удаления", Toast.LENGTH_SHORT).show()
                }
            }
        }
    }
}
