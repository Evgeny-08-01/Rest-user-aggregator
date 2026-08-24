package com.example.sub

data class Subscription(
    val id: Int,
    val serviceName: String,
    val price: Int,
    val userId: String,
    val startDate: String,
    val endDate: String = ""
)
