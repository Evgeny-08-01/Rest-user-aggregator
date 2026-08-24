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

class LoginActivity : AppCompatActivity() {

    private lateinit var etEmail: EditText
    private lateinit var etPassword: EditText
    private lateinit var btnLogin: Button
    private lateinit var tvError: TextView

    private val scope = CoroutineScope(Dispatchers.Main)

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_login)

        etEmail = findViewById(R.id.etLoginEmail)
        etPassword = findViewById(R.id.etLoginPassword)
        val btnToRegister: Button = findViewById(R.id.btnToRegister)
        btnLogin = findViewById(R.id.btnLogin)
        tvError = findViewById(R.id.tvLoginError)

        // Если уже есть токен — идём в главный экран
        if (TokenManager.getToken(this).isNotBlank()) {
            startMainActivity()
            return
        }

        // Кнопка "Зарегистрироваться" — открывает экран регистрации
        btnToRegister.setOnClickListener {
            startActivity(android.content.Intent(this, RegisterActivity::class.java))
        }

        // Кнопка "Войти"
        btnLogin.setOnClickListener {
            val email = etEmail.text.toString().trim()
            val password = etPassword.text.toString().trim()

            if (email.isBlank() || password.isBlank()) {
                showError("Введите email и пароль")
                return@setOnClickListener
            }

            btnLogin.isEnabled = false
            tvError.visibility = android.view.View.GONE

            scope.launch {
                val success = ApiService.login(this@LoginActivity, email, password)
                withContext(Dispatchers.Main) {
                    btnLogin.isEnabled = true
                    if (success) {
                        startMainActivity()
                    } else {
                        showError("Неверный email или пароль")
                    }
                }
            }
        }
    }

    private fun showError(msg: String) {
        tvError.text = msg
        tvError.visibility = android.view.View.VISIBLE
    }

    private fun startMainActivity() {
        startActivity(android.content.Intent(this, MainActivity::class.java))
        finish()
    }
}
