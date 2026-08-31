package com.example.sub

import android.os.Bundle
import android.widget.Button
import android.widget.EditText
import android.widget.TextView
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.recyclerview.widget.LinearLayoutManager
import androidx.recyclerview.widget.RecyclerView
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

class MainActivity : AppCompatActivity() {

    private lateinit var recyclerView: RecyclerView
    private lateinit var adapter: SubscriptionAdapter
    private lateinit var tvTotal: TextView
    private lateinit var tvUserName: TextView
    private lateinit var etStartDate: EditText
    private lateinit var etEndDate: EditText
    private lateinit var btnAdd: Button
    private lateinit var btnTotal: Button
    private lateinit var btnLogout: Button

    private val scope = CoroutineScope(Dispatchers.Main)

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        recyclerView = findViewById(R.id.recyclerView)
        tvTotal = findViewById(R.id.tvTotal)
        tvUserName = findViewById(R.id.tvUserName)
        etStartDate = findViewById(R.id.etStartDate)
        etEndDate = findViewById(R.id.etEndDate)
        btnAdd = findViewById(R.id.btnAdd)
        btnTotal = findViewById(R.id.btnTotal)
        val btnRefresh: Button = findViewById(R.id.btnRefresh)
        btnLogout = findViewById(R.id.btnLogout)

        val user = getUser(this)
        tvUserName.text = "👤 ${user?.email ?: "Пользователь"}"

        adapter = SubscriptionAdapter(
            items = emptyList(),
            onEdit = { sub ->
                startActivity(android.content.Intent(this, EditSubscriptionActivity::class.java)
                    .putExtra("subscription_id", sub.id))
            },
            onDelete = { sub -> deleteSubscription(sub.id) }
        )

        recyclerView.layoutManager = LinearLayoutManager(this)
        recyclerView.adapter = adapter

        val currentUser = getUser(this)
        if (currentUser?.role == "admin") {
            val btnTemplates = Button(this).apply {
                text = "📋 Управление шаблонами"
                setBackgroundColor(android.graphics.Color.parseColor("#16a34a"))
                setTextColor(android.graphics.Color.WHITE)
                setOnClickListener {
                    startActivity(android.content.Intent(this@MainActivity, TemplatesActivity::class.java))
                }
            }
            val parent = btnAdd.parent as? android.view.ViewGroup
            parent?.addView(btnTemplates, parent.indexOfChild(btnAdd))
        }

        btnAdd.setOnClickListener {
            startActivity(android.content.Intent(this, AddSubscriptionActivity::class.java))
        }

        btnRefresh.setOnClickListener {
            loadSubscriptions()
            showToast("Обновлено")
        }

        btnLogout.setOnClickListener {
            TokenManager.saveToken(this, "")
            startActivity(android.content.Intent(this, LoginActivity::class.java))
            finish()
        }

        btnTotal.setOnClickListener {
            val start = etStartDate.text.toString()
            val end = etEndDate.text.toString()
            if (start.isBlank() || end.isBlank()) {
                showToast("Введите обе даты (MM-YYYY)")
                return@setOnClickListener
            }
            scope.launch {
                val total = ApiService.getTotalCost(this@MainActivity, start, end)
                withContext(Dispatchers.Main) {
                    if (total != null) {
                        tvTotal.text = "💰 Сумма: $total ₽"
                    } else {
                        showToast("Ошибка расчёта")
                    }
                }
            }
        }

        loadSubscriptions()
    }

    override fun onResume() {
        super.onResume()
        loadSubscriptions()
    }

    private fun loadSubscriptions() {
    scope.launch {
        try {
            val restList = ApiService.getSubscriptions(this@MainActivity)
            withContext(Dispatchers.Main) {
                adapter.updateData(restList)
                if (restList.isNotEmpty()) {
                    showToast("✅ REST загрузил ${restList.size} подписок")
                } else {
                    showToast("📭 Нет подписок")
                }
            }
        } catch (e: Exception) {
            withContext(Dispatchers.Main) {
                showToast("❌ Ошибка загрузки: ${e.message}")
            }
        }
    }
}
    private fun deleteSubscription(id: Int) {
        scope.launch {
            val success = ApiService.deleteSubscription(this@MainActivity, id)
            withContext(Dispatchers.Main) {
                if (success) {
                    showToast("Удалено")
                    loadSubscriptions()
                } else {
                    showToast("Ошибка удаления")
                }
            }
        }
    }

    private fun showToast(msg: String) {
        Toast.makeText(this, msg, Toast.LENGTH_SHORT).show()
    }
}
