package com.example.app;

import java.util.List;
import java.util.Map;
import com.example.service.UserService;
import static org.junit.Assert.assertEquals;

public class Main {
    public static void main(String[] args) {
        System.out.println("Hello");
    }
}

interface AppConfig {
    String getName();
}
