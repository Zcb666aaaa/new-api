/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useEffect, useRef, useCallback } from 'react';
import { Modal, Spin, Typography } from '@douyinfe/semi-ui';
import { CreditCard } from 'lucide-react';

const { Text } = Typography;

/**
 * PaymentIframeModal
 *
 * 在弹窗内嵌 iframe 完成支付，避免跳转到新页面。
 * 原理：动态创建隐藏 <form>，将 target 指向 iframe，POST 提交后
 * 第三方收银台页面在 iframe 内加载，用户可在弹窗内扫码/完成支付。
 *
 * Props:
 *   open        — 是否显示
 *   onClose     — 关闭回调
 *   payUrl      — POST 目标地址（来自后端 url 字段）
 *   payParams   — POST 参数（来自后端 data 字段，key-value 对象）
 *   t           — i18n 函数
 */
const PaymentIframeModal = ({ open, onClose, payUrl, payParams, t }) => {
  const iframeRef = useRef(null);
  const formRef = useRef(null);
  // 是否已经提交过（避免重复提交）
  const submittedRef = useRef(false);

  const submitToIframe = useCallback(() => {
    if (!payUrl || !payParams || submittedRef.current) return;
    if (!iframeRef.current) return;

    submittedRef.current = true;

    // 创建隐藏 form，target 指向 iframe name
    const form = document.createElement('form');
    form.action = payUrl;
    form.method = 'POST';
    form.target = 'payment-iframe';
    form.style.display = 'none';

    for (const key in payParams) {
      const input = document.createElement('input');
      input.type = 'hidden';
      input.name = key;
      input.value = payParams[key];
      form.appendChild(input);
    }

    document.body.appendChild(form);
    formRef.current = form;
    form.submit();
    document.body.removeChild(form);
  }, [payUrl, payParams]);

  useEffect(() => {
    if (open) {
      submittedRef.current = false;
    }
  }, [open, payUrl, payParams]);

  // iframe 加载完毕后提交表单
  const handleIframeLoad = useCallback(() => {
    submitToIframe();
  }, [submitToIframe]);

  const handleClose = () => {
    submittedRef.current = false;
    onClose();
  };

  return (
    <Modal
      title={
        <div className='flex items-center'>
          <CreditCard className='mr-2' size={18} />
          {t('在线支付')}
        </div>
      }
      visible={open}
      onCancel={handleClose}
      footer={null}
      maskClosable={false}
      width={820}
      centered
      bodyStyle={{ padding: 0 }}
    >
      <div className='flex flex-col' style={{ height: 600 }}>
        <div className='px-4 py-2 border-b border-slate-100 dark:border-slate-700'>
          <Text size='small' className='text-slate-500 dark:text-slate-400'>
            {t('请在下方完成支付，支付成功后页面将自动更新')}
          </Text>
        </div>
        <div className='relative flex-1'>
          {!payUrl && (
            <div className='absolute inset-0 flex items-center justify-center'>
              <Spin size='large' />
            </div>
          )}
          {/* iframe name 与 form target 保持一致 */}
          <iframe
            ref={iframeRef}
            name='payment-iframe'
            title={t('在线支付')}
            onLoad={handleIframeLoad}
            style={{
              width: '100%',
              height: '100%',
              border: 'none',
              display: payUrl ? 'block' : 'none',
            }}
            sandbox='allow-forms allow-scripts allow-same-origin allow-top-navigation allow-popups allow-popups-to-escape-sandbox'
          />
        </div>
      </div>
    </Modal>
  );
};

export default PaymentIframeModal;
