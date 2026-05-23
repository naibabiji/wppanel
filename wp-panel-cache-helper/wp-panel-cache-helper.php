<?php
/**
 * Plugin Name: WP Panel Cache Helper
 * Plugin URI:  https://github.com/naibabiji/wp-panel
 * Description: 与 WP Panel 面板配合，一键清除 Nginx FastCGI 缓存。发布/更新文章自动清除。
 * Version:     1.0.0
 * Author:      WP Panel
 * License:     GPL-2.0+
 */

if (!defined('ABSPATH')) exit;

register_uninstall_hook(__FILE__, 'wpp_cache_uninstall');
function wpp_cache_uninstall() {
    delete_option('wpp_cache_ttl');
    delete_option('wpp_cache_verified');
    delete_option('wpp_cache_log');
}

class WP_Panel_Cache_Helper {

    const OPTION_CACHE_TTL  = 'wpp_cache_ttl';
    const OPTION_VERIFIED   = 'wpp_cache_verified';
    const OPTION_CACHE_LOG  = 'wpp_cache_log';

    private static function load_config() {
        $file = __DIR__ . '/wp-panel-config.json';
        if (!file_exists($file)) return null;
        return json_decode(file_get_contents($file), true);
    }

    private static function get_panel_url() {
        $cfg = self::load_config();
        return $cfg ? $cfg['panel_url'] : '';
    }

    private static function get_api_key() {
        $cfg = self::load_config();
        return $cfg ? $cfg['api_key'] : '';
    }

    public static function init() {
        add_action('admin_bar_menu', [__CLASS__, 'admin_bar_button'], 100);
        add_action('admin_menu', [__CLASS__, 'settings_page']);
        add_action('admin_post_wpp_cache_clear', [__CLASS__, 'handle_clear']);
        add_action('save_post', [__CLASS__, 'auto_clear'], 99, 1);
        add_action('deleted_post', [__CLASS__, 'auto_clear'], 99, 1);
        add_action('wp_update_comment_count', [__CLASS__, 'auto_comment_clear']);
        add_filter('plugin_action_links_' . plugin_basename(__FILE__), [__CLASS__, 'action_links']);
        add_action('admin_notices', [__CLASS__, 'clear_notice']);
    }

    public static function action_links($links) {
        $links[] = '<a href="' . admin_url('options-general.php?page=wp-panel-cache') . '">设置</a>';
        return $links;
    }

    public static function settings_page() {
        add_options_page('WP Panel Cache', 'WP Panel Cache', 'manage_options', 'wp-panel-cache', [__CLASS__, 'render_settings']);
    }

    public static function render_settings() {
        $cfg = self::load_config();
        $panelUrl = self::get_panel_url();
        $apiKey = self::get_api_key();

        $notice = '';
        if (isset($_POST['wpp_save'])) {
            check_admin_referer('wpp_cache_settings');
            $ttl = intval($_POST['cache_ttl']);
            if ($ttl >= 10 && $ttl <= 86400) {
                update_option(self::OPTION_CACHE_TTL, $ttl);
                self::update_cache_settings($ttl);
            }
            $notice = '<div class="notice notice-success"><p>设置已保存</p></div>';
        }

        $currentDomain = parse_url(home_url(), PHP_URL_HOST);
        $ttl = get_option(self::OPTION_CACHE_TTL, '300');
        $missing = !$panelUrl || !$apiKey;
        $log = get_option(self::OPTION_CACHE_LOG, []);
        ?>
        <div class="wrap">
            <h1>WP Panel Cache Helper</h1>
            <p>由 <a href="https://github.com/naibabiji/wp-panel" target="_blank">WP Panel</a> 面板统一管理。当前站点：<code><?php echo esc_html($currentDomain); ?></code></p>
            <?php echo $notice; ?>
            <?php if ($missing): ?>
                <div class="notice notice-error"><p><strong>配置文件缺失</strong> — 请在 WP Panel 面板中进入该网站详情页，点击 FastCGI 缓存卡片的「安装配套插件」按钮完成初始化。</p></div>
            <?php endif; ?>
            <div id="wpp-verify-msg"></div>
            <hr>
            <table class="form-table">
                <tr>
                    <th>面板地址</th>
                    <td>
                        <code><?php echo esc_html($panelUrl ?: '未配置'); ?></code>
                        <p class="description">面板自动写入，无需手动填写。</p>
                    </td>
                </tr>
                <tr>
                    <th>API Key</th>
                    <td>
                        <code><?php echo esc_html($apiKey ? substr($apiKey, 0, 8) . '...' : '未配置'); ?></code>
                        <p class="description">每个站点独立密钥，面板安装插件时自动生成。</p>
                    </td>
                </tr>
                <tr>
                    <th><label for="wpp-cache-ttl">缓存有效期（秒）</label></th>
                    <td>
                        <input id="wpp-cache-ttl" name="cache_ttl" type="number" class="regular-text" value="<?php echo esc_attr($ttl); ?>" min="10" max="86400" form="wpp-form">
                        <p class="description">建议 300-3600 秒（5分钟到1小时）。修改后 Nginx 自动重载。</p>
                    </td>
                </tr>
            </table>
            <form id="wpp-form" method="post">
                <?php wp_nonce_field('wpp_cache_settings'); ?>
                <p>
                    <button type="submit" name="wpp_save" class="button button-primary">保存设置</button>
                    <button type="button" id="wpp-verify-btn" class="button">验证连接</button>
                </p>
            </form>

            <?php if (!empty($log)): ?>
            <hr>
            <h2>最近清除记录</h2>
            <table class="wp-list-table widefat fixed striped" style="max-width:600px">
                <thead>
                    <tr>
                        <th>时间</th>
                        <th>方式</th>
                        <th>结果</th>
                    </tr>
                </thead>
                <tbody>
                    <?php foreach ($log as $entry): ?>
                    <tr>
                        <td><?php echo esc_html($entry['time']); ?></td>
                        <td><?php echo $entry['type'] === 'manual' ? '手动清除' : '自动清除（发布文章）'; ?></td>
                        <td><?php echo !empty($entry['success']) ? '<span style="color:green">成功</span>' : '<span style="color:red">失败</span>'; ?></td>
                    </tr>
                    <?php endforeach; ?>
                </tbody>
            </table>
            <?php endif; ?>

            <script>
            document.getElementById('wpp-verify-btn').addEventListener('click', function() {
                var btn = this, msg = document.getElementById('wpp-verify-msg');
                btn.disabled = true;
                btn.textContent = '验证中...';
                fetch('<?php echo admin_url('admin-ajax.php'); ?>?action=wpp_verify_connection&_wpnonce=<?php echo wp_create_nonce('wpp_cache_settings'); ?>')
                    .then(r => r.json())
                    .then(data => {
                        if (data.success) {
                            msg.innerHTML = '<div class="notice notice-success"><p>✓ 连接成功 — 面板 API 响应正常</p></div>';
                        } else {
                            msg.innerHTML = '<div class="notice notice-error"><p>✗ 连接失败：' + (data.data?.message || '未知错误') + '</p></div>';
                        }
                    })
                    .catch(e => {
                        msg.innerHTML = '<div class="notice notice-error"><p>✗ 网络错误：无法连接到面板 (' + e.message + ')</p></div>';
                    })
                    .finally(() => { btn.disabled = false; btn.textContent = '验证连接'; });
            });
            </script>
        </div>
        <?php
    }

    public static function clear_notice() {
        if (!isset($_GET['wpp_cleared'])) return;
        if ($_GET['wpp_cleared'] === '1') {
            echo '<div class="notice notice-success is-dismissible"><p>Nginx 缓存已清除，旧页面将在几分钟内更新。</p></div>';
        } else {
            echo '<div class="notice notice-error is-dismissible"><p>清除缓存失败，请检查面板连接是否正常。</p></div>';
        }
    }

    public static function admin_bar_button($bar) {
        $panelUrl = self::get_panel_url();
        if (!$panelUrl) return;
        $bar->add_node([
            'id'    => 'wpp-clear-cache',
            'title' => '清除 Nginx 缓存',
            'href'  => wp_nonce_url(admin_url('admin-post.php?action=wpp_cache_clear'), 'wpp_cache_clear'),
        ]);
    }

    public static function handle_clear() {
        check_admin_referer('wpp_cache_clear');
        $resp = self::do_clear();
        $success = !empty($resp['success']);
        self::log_clear('manual', $success);
        wp_redirect(add_query_arg('wpp_cleared', $success ? '1' : '0', wp_get_referer() ?: admin_url()));
        exit;
    }

    public static function auto_clear($post_id) {
        // 跳过自动草稿和修订版本
        if (wp_is_post_revision($post_id) || wp_is_post_autosave($post_id)) return;
        $post = get_post($post_id);
        if (!$post || in_array($post->post_status, ['draft', 'auto-draft', 'inherit'])) return;
        if (!in_array($post->post_status, ['publish', 'trash', 'future', 'private'])) return;

        $resp = self::do_clear();
        $success = !empty($resp['success']);
        self::log_clear('auto', $success);
    }

    public static function auto_comment_clear($_) {
        $resp = self::do_clear();
        $success = !empty($resp['success']);
        self::log_clear('auto', $success);
    }

    private static function log_clear($type, $success) {
        $log = get_option(self::OPTION_CACHE_LOG, []);
        array_unshift($log, [
            'time'    => current_time('Y-m-d H:i:s'),
            'type'    => $type,
            'success' => $success,
        ]);
        update_option(self::OPTION_CACHE_LOG, array_slice($log, 0, 10));
    }

    private static function update_cache_settings($ttl) {
        $domain = parse_url(home_url(), PHP_URL_HOST);
        self::api_request('PUT', '/api/sites/cache-settings', ['domain' => $domain, 'ttl' => $ttl]);
    }

    private static function do_clear() {
        $domain = parse_url(home_url(), PHP_URL_HOST);
        $resp = self::api_request('DELETE', '/api/sites/clear-cache', ['domain' => $domain]);
        $data = json_decode($resp, true);
        return ['success' => !empty($data['success']), 'message' => $data['message'] ?? ''];
    }

    public static function api_request_public($method, $path, $body = null) {
        return self::api_request($method, $path, $body);
    }

    private static function api_request($method, $path, $body = null) {
        $baseUrl = self::get_panel_url();
        $apiKey  = self::get_api_key();
        if (!$baseUrl || !$apiKey) return false;

        $args = [
            'method'    => $method,
            'headers'   => [
                'X-WP-Panel-Key' => $apiKey,
                'Content-Type'   => 'application/json',
            ],
            'timeout'   => 10,
            'sslverify' => false,
        ];

        if ($body) {
            $args['body'] = json_encode($body);
        }

        $response = wp_remote_request($baseUrl . $path, $args);
        if (is_wp_error($response)) return false;
        return wp_remote_retrieve_body($response);
    }
}

add_action('init', ['WP_Panel_Cache_Helper', 'init']);

add_action('wp_ajax_wpp_verify_connection', function() {
    check_ajax_referer('wpp_cache_settings');
    $domain = parse_url(home_url(), PHP_URL_HOST);
    $resp = WP_Panel_Cache_Helper::api_request_public('GET', '/api/sites/find?domain=' . urlencode($domain));
    if (!$resp) {
        wp_send_json(['success' => false, 'data' => ['message' => '无响应，请检查面板地址']]);
        return;
    }
    $data = json_decode($resp, true);
    if (!empty($data['success'])) {
        update_option('wpp_cache_verified', '1');
        wp_send_json(['success' => true, 'data' => ['message' => '连接成功']]);
    } else {
        wp_send_json(['success' => false, 'data' => ['message' => $data['message'] ?? 'API 返回错误']]);
    }
});
