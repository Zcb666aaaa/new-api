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

import React, { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Modal,
  Form,
  Space,
  InputNumber,
  Tag,
} from '@douyinfe/semi-ui';
import { IconDelete, IconPlus, IconEdit, IconSave } from '@douyinfe/semi-icons';
import { API, showError, showSuccess } from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function TieredPricingEditor(props) {
  const { t } = useTranslation();
  const [models, setModels] = useState([]);
  const [visible, setVisible] = useState(false);
  const [currentModel, setCurrentModel] = useState(null);
  const [currentTiers, setCurrentTiers] = useState([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    try {
      const tieredPrice = JSON.parse(props.options.ModelTieredPrice || '{}');
      const modelData = Object.entries(tieredPrice).map(([name, config]) => ({
        name,
        tiers: config.tiers || [],
      }));
      setModels(modelData);
    } catch (error) {
      console.error('JSON解析错误:', error);
    }
  }, [props.options]);

  const SubmitData = async () => {
    setLoading(true);
    const output = {};

    try {
      models.forEach((model) => {
        output[model.name] = {
          tiers: model.tiers,
        };
      });

      const res = await API.put('/api/option/', {
        key: 'ModelTieredPrice',
        value: JSON.stringify(output, null, 2),
      });

      if (res.data.success) {
        showSuccess(t('保存成功'));
        props.refresh();
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      console.error('保存失败:', error);
      showError(t('保存失败，请重试'));
    } finally {
      setLoading(false);
    }
  };

  const addTier = () => {
    setCurrentTiers([
      ...currentTiers,
      {
        min_tokens: 0,
        max_tokens: -1,
        input: 0,
        output: 0,
      },
    ]);
  };

  const updateTier = (index, field, value) => {
    const updated = [...currentTiers];
    updated[index][field] = value;
    setCurrentTiers(updated);
  };

  const deleteTier = (index) => {
    setCurrentTiers(currentTiers.filter((_, i) => i !== index));
  };

  const addOrUpdateModel = (modelName) => {
    if (!modelName || currentTiers.length === 0) {
      showError(t('请填写模型名称并至少添加一个阶梯'));
      return;
    }

    const existingIndex = models.findIndex((m) => m.name === modelName);
    if (existingIndex >= 0) {
      // Update existing
      setModels((prev) =>
        prev.map((m, i) =>
          i === existingIndex ? { name: modelName, tiers: currentTiers } : m,
        ),
      );
      showSuccess(t('更新成功'));
    } else {
      // Add new
      setModels([...models, { name: modelName, tiers: currentTiers }]);
      showSuccess(t('添加成功'));
    }

    setVisible(false);
    setCurrentModel(null);
    setCurrentTiers([]);
  };

  const editModel = (model) => {
    setCurrentModel(model.name);
    setCurrentTiers(model.tiers);
    setVisible(true);
  };

  const deleteModel = (name) => {
    setModels(models.filter((m) => m.name !== name));
  };

  const columns = [
    {
      title: t('模型名称'),
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: t('阶梯数量'),
      dataIndex: 'tiers',
      key: 'tiers',
      render: (tiers) => (
        <Tag color='blue'>{tiers ? tiers.length : 0} {t('个阶梯')}</Tag>
      ),
    },
    {
      title: t('操作'),
      key: 'action',
      render: (_, record) => (
        <Space>
          <Button
            type='primary'
            icon={<IconEdit />}
            onClick={() => editModel(record)}
          />
          <Button
            icon={<IconDelete />}
            type='danger'
            onClick={() => deleteModel(record.name)}
          />
        </Space>
      ),
    },
  ];

  const tierColumns = [
    {
      title: t('最小Token数'),
      dataIndex: 'min_tokens',
      key: 'min_tokens',
      render: (text, record, index) => (
        <InputNumber
          value={text}
          onChange={(value) => updateTier(index, 'min_tokens', value)}
          placeholder={t('0表示无限制')}
          style={{ width: '100%' }}
        />
      ),
    },
    {
      title: t('最大Token数'),
      dataIndex: 'max_tokens',
      key: 'max_tokens',
      render: (text, record, index) => (
        <InputNumber
          value={text}
          onChange={(value) => updateTier(index, 'max_tokens', value)}
          placeholder={t('-1表示无限制')}
          style={{ width: '100%' }}
        />
      ),
    },
    {
      title: t('输入价格 ($/1M tokens)'),
      dataIndex: 'input',
      key: 'input',
      render: (text, record, index) => (
        <InputNumber
          value={text}
          onChange={(value) => updateTier(index, 'input', value)}
          step={0.01}
          style={{ width: '100%' }}
        />
      ),
    },
    {
      title: t('输出价格 ($/1M tokens)'),
      dataIndex: 'output',
      key: 'output',
      render: (text, record, index) => (
        <InputNumber
          value={text}
          onChange={(value) => updateTier(index, 'output', value)}
          step={0.01}
          style={{ width: '100%' }}
        />
      ),
    },
    {
      title: t('操作'),
      key: 'action',
      render: (_, record, index) => (
        <Button
          icon={<IconDelete />}
          type='danger'
          onClick={() => deleteTier(index)}
        />
      ),
    },
  ];

  return (
    <>
      <Space vertical align='start' style={{ width: '100%' }}>
        <Space className='mt-2'>
          <Button
            icon={<IconPlus />}
            onClick={() => {
              setCurrentModel(null);
              setCurrentTiers([]);
              setVisible(true);
            }}
          >
            {t('添加阶梯计费模型')}
          </Button>
          <Button
            type='primary'
            icon={<IconSave />}
            onClick={SubmitData}
            loading={loading}
          >
            {t('保存更改')}
          </Button>
        </Space>
        <Table columns={columns} dataSource={models} />
      </Space>

      <Modal
        title={currentModel ? t('编辑阶梯计费') : t('添加阶梯计费')}
        visible={visible}
        onCancel={() => {
          setVisible(false);
          setCurrentModel(null);
          setCurrentTiers([]);
        }}
        onOk={() => addOrUpdateModel(currentModel)}
        width={900}
      >
        <Form>
          <Form.Input
            field='name'
            label={t('模型名称')}
            value={currentModel || ''}
            onChange={(value) => setCurrentModel(value)}
            placeholder='claude-3-opus'
            disabled={!!currentModel && models.some((m) => m.name === currentModel)}
          />

          <div style={{ marginTop: 16, marginBottom: 8 }}>
            <Space>
              <strong>{t('阶梯配置')}</strong>
              <Button icon={<IconPlus />} size='small' onClick={addTier}>
                {t('添加阶梯')}
              </Button>
            </Space>
          </div>

          <Table
            columns={tierColumns}
            dataSource={currentTiers}
            pagination={false}
          />

          <div style={{ marginTop: 16, padding: 12, background: '#f8f9fa', borderRadius: 4 }}>
            <strong>{t('说明')}</strong>
            <ul style={{ margin: '8px 0 0 20px', paddingLeft: 0 }}>
              <li>{t('最小Token数：该阶梯的起始Token数（包含），0表示从0开始')}</li>
              <li>{t('最大Token数：该阶梯的结束Token数（不包含），-1表示无限制')}</li>
              <li>{t('价格单位：美元/百万Token')}</li>
              <li>{t('示例：输入32000 tokens，会查找min_tokens <= 32000 < max_tokens的阶梯')}</li>
            </ul>
          </div>
        </Form>
      </Modal>
    </>
  );
}
