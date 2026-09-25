plugins {
    id("com.android.application")
    // The Flutter Gradle Plugin must be applied after the Android and Kotlin Gradle plugins.
    id("dev.flutter.flutter-gradle-plugin")
}

android {
    namespace = "com.salesorder.sales_order_app"
    compileSdk = 37 // flutter_secure_storage 11 需 37;高於 AGP 9.1 建議值,以 gradle.properties 抑制警告
    ndkVersion = flutter.ndkVersion

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    defaultConfig {
        // App Links 的網域（規格 §4.2 深層連結）。真實網域未定案，故集中在此一處：
        // 改這裡就同時更新 AndroidManifest 的兩個 intent-filter host，兩平台再各自對齊
        // （iOS 的另一半在 `ios/Runner/Runner.entitlements`，部署面在 `/.well-known/`）。
        //
        // 為何要 host 而非自訂 scheme：任何 App 都能宣告 `salesorder://`，不具網域歸屬
        // 保證；https + autoVerify 才由作業系統向網域驗證所有權（assetlinks.json）。
        manifestPlaceholders["appLinkHost"] = providers.gradleProperty("appLinkHost")
            .orElse("app.salesorder.example.com").get()

        // TODO: Specify your own unique Application ID (https://developer.android.com/studio/build/application-id.html).
        applicationId = "com.salesorder.sales_order_app"
        // You can update the following values to match your application needs.
        // For more information, see: https://flutter.dev/to/review-gradle-config.
        minSdk = flutter.minSdkVersion
        targetSdk = flutter.targetSdkVersion
        // Uses the version code from pubspec.yaml. When using split APKs, 1000 * ABI_VERSION
        // is added automatically by Flutter. (https://developer.android.com/studio/build/configure-apk-splits#configure-APK-versions)
        // You can force using the value of versionCode by specifying the `-P force-version-code-ignoring-abi=true`
        // flag during build.
        versionCode = flutter.versionCode
        versionName = flutter.versionName
    }

    // dev / prod flavor(D29);入口分別為 lib/main_dev.dart / lib/main_prod.dart。
    flavorDimensions += "env"
    productFlavors {
        create("dev") {
            dimension = "env"
            applicationIdSuffix = ".dev"
            versionNameSuffix = "-dev"
        }
        create("prod") {
            dimension = "env"
        }
    }

    buildTypes {
        release {
            // TODO: Add your own signing config for the release build.
            // Signing with the debug keys for now, so `flutter run --release` works.
            signingConfig = signingConfigs.getByName("debug")
        }
    }
}

kotlin {
    compilerOptions {
        jvmTarget = org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17
    }
}

flutter {
    source = "../.."
}
