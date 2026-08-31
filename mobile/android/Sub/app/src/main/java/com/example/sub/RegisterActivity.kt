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

class RegisterActivity : AppCompatActivity() {

    private lateinit var etEmail: EditText
    private lateinit var etPassword: EditText
    private lateinit var btnRegister: Button
    private lateinit var btnToLogin: Button
    private lateinit var tvError: TextView

    private val scope = CoroutineScope(Dispatchers.Main)

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_register)

        etEmail = findViewById(R.id.etRegEmail)
        etPassword = findViewById(R.id.etRegPassword)
        btnRegister = findViewById(R.id.btnRegister)
        btnToLogin = findViewById(R.id.btnToLogin)
        tvError = findViewById(R.id.tvRegError)

        // Подсказка для роли

        btnRegister.setOnClickListener {
            val email = etEmail.text.toString().trim()
            val password = etPassword.text.toString().trim()

            if (email.isBlank() || password.isBlank()) {
                showError("Заполните все поля")
                return@setOnClickListener
            }

            if (password.length < 6) {
                showError("Пароль должен быть минимум 6 символов")
                return@setOnClickListener
            }

            btnRegister.isEnabled = false
            tvError.visibility = android.view.View.GONE

            scope.launch {
               val success = ApiService.register(this@RegisterActivity, email, password)
                withContext(Dispatchers.Main) {
                    btnRegister.isEnabled = true
                    if (success) {
                        Toast.makeText(this@RegisterActivity, "✅ Регистрация успешна! Войдите.", Toast.LENGTH_LONG).show()
                        finish()
                    } else {
                        showError("Ошибка регистрации. Возможно, email уже занят.")
                    }
                }
            }
        }

        btnToLogin.setOnClickListener {
            finish()
        }
    }

    private fun showError(msg: String) {
        tvError.text = msg
        tvError.visibility = android.view.View.VISIBLE
    }
}
