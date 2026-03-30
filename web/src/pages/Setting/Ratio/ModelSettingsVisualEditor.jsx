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

import React, { useEffect, useState, useRef } from 'react';
import {
  Table,
  Button,
  Input,
  InputNumber,
  Modal,
  Form,
  Space,
  RadioGroup,
  Radio,
  Checkbox,
  Tag,
  Divider,
  Typography,
  Card,
  Avatar,
  Banner,
} from '@douyinfe/semi-ui';

const { Text } = Typography;
import {
  IconDelete,
  IconPlus,
  IconMinus,
  IconSearch,
  IconSave,
  IconEdit,
  IconLayers,
  IconAlertTriangle,
} from '@douyinfe/semi-icons';
import { FileText, DollarSign, Layers } from 'lucide-react';
import { API, showError, showSuccess, getQuotaPerUnit } from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function ModelSettingsVisualEditor(props) {
  const { t } = useTranslation();
  const [models, setModels] = useState([]);
  const [visible, setVisible] = useState(false);
  const [isEditMode, setIsEditMode] = useState(false);
  const [currentModel, setCurrentModel] = useState(null);
  const [searchText, setSearchText] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [pricingMode, setPricingMode] = useState('per-token'); // 'per-token', 'per-request', 'per-second', 'per-tiered'
  const [pricingSubMode, setPricingSubMode] = useState('ratio'); // 'ratio' or 'token-price'
  const [conflictOnly, setConflictOnly] = useState(false);
  const [editingTiers, setEditingTiers] = useState([]); // 阶梯计费：当前编辑中的阶梯列表
  // 分组定价相关状态
  const [groupPricingVisible, setGroupPricingVisible] = useState(false);
  const [groupPricingModel, setGroupPricingModel] = useState(null); // 当前编辑的模型名
  const [groupPricingRows, setGroupPricingRows] = useState([]); // [{group, value}]
  const [groupPricingSaving, setGroupPricingSaving] = useState(false);
  const [groupPricingGlobalTiers, setGroupPricingGlobalTiers] = useState([]); // per-tiered 全局阶梯参考
  const formRef = useRef(null);
  const pageSize = 10;
  const quotaPerUnit = getQuotaPerUnit();

  useEffect(() => {
    try {
      const modelPrice = JSON.parse(props.options.ModelPrice || '{}');
      const modelPricePerSecond = JSON.parse(
        props.options.ModelPricePerSecond || '{}',
      );
      const modelRatio = JSON.parse(props.options.ModelRatio || '{}');
      const completionRatio = JSON.parse(props.options.CompletionRatio || '{}');
      const modelTieredPrice = JSON.parse(
        props.options.ModelTieredPrice || '{}',
      );
      const groupModelRatio = JSON.parse(props.options.GroupModelRatio || '{}');

      // 合并所有模型名称（含阶梯计费模型）
      const modelNames = new Set([
        ...Object.keys(modelPrice),
        ...Object.keys(modelPricePerSecond),
        ...Object.keys(modelRatio),
        ...Object.keys(completionRatio),
        ...Object.keys(modelTieredPrice),
      ]);

      const modelData = Array.from(modelNames).map((name) => {
        const perSecondPrice =
          modelPricePerSecond[name] === undefined ? '' : modelPricePerSecond[name];
        const regularPrice = modelPrice[name] === undefined ? '' : modelPrice[name];
        const ratio = modelRatio[name] === undefined ? '' : modelRatio[name];
        const comp =
          completionRatio[name] === undefined ? '' : completionRatio[name];
        const tiered = modelTieredPrice[name]; // { tiers: [...] } or undefined

        // 检查是否有分组定价
        const groupPricingGroups = [];
        for (const [group, modelMap] of Object.entries(groupModelRatio)) {
          if (modelMap && modelMap[name] !== undefined) {
            groupPricingGroups.push(group);
          }
        }

        // Determine billing mode: tiered > per-second > per-request > per-token
        let billingMode;
        if (tiered) {
          billingMode = 'per-tiered';
        } else if (perSecondPrice !== '') {
          billingMode = 'per-second';
        } else if (regularPrice !== '') {
          billingMode = 'per-request';
        } else {
          billingMode = 'per-token';
        }

        const price =
          billingMode === 'per-second' ? perSecondPrice : regularPrice;

        return {
          name,
          price,
          ratio,
          completionRatio: comp,
          billingMode,
          tiers: tiered ? tiered.tiers || [] : [],
          hasConflict: price !== '' && (ratio !== '' || comp !== ''),
          groupPricingGroups, // 存储有分组定价的分组列表
        };
      });

      setModels(modelData);
    } catch (error) {
      console.error('JSON解析错误:', error);
    }
  }, [props.options]);

  // 首先声明分页相关的工具函数
  const getPagedData = (data, currentPage, pageSize) => {
    const start = (currentPage - 1) * pageSize;
    const end = start + pageSize;
    return data.slice(start, end);
  };

  // 在 return 语句之前，先处理过滤和分页逻辑
  const filteredModels = models.filter((model) => {
    const keywordMatch = searchText ? model.name.includes(searchText) : true;
    const conflictMatch = conflictOnly ? model.hasConflict : true;
    return keywordMatch && conflictMatch;
  });

  // 然后基于过滤后的数据计算分页数据
  const pagedData = getPagedData(filteredModels, currentPage, pageSize);

  const SubmitData = async () => {
    setLoading(true);
    const output = {
      ModelPrice: {},
      ModelPricePerSecond: {},
      ModelRatio: {},
      CompletionRatio: {},
      ModelTieredPrice: {},
    };

    try {
      models.forEach((model) => {
        if (model.billingMode === 'per-tiered') {
          // 阶梯计费：只写 ModelTieredPrice
          if (model.tiers && model.tiers.length > 0) {
            output.ModelTieredPrice[model.name] = { tiers: model.tiers };
          }
        } else if (model.billingMode === 'per-second' && model.price !== '') {
          output.ModelPricePerSecond[model.name] = parseFloat(model.price);
        } else if (model.billingMode === 'per-request' && model.price !== '') {
          output.ModelPrice[model.name] = parseFloat(model.price);
        } else {
          if (model.ratio !== '')
            output.ModelRatio[model.name] = parseFloat(model.ratio);
          if (model.completionRatio !== '')
            output.CompletionRatio[model.name] = parseFloat(model.completionRatio);
        }
      });

      const finalOutput = {
        ModelPrice: JSON.stringify(output.ModelPrice, null, 2),
        ModelPricePerSecond: JSON.stringify(output.ModelPricePerSecond, null, 2),
        ModelRatio: JSON.stringify(output.ModelRatio, null, 2),
        CompletionRatio: JSON.stringify(output.CompletionRatio, null, 2),
        ModelTieredPrice: JSON.stringify(output.ModelTieredPrice, null, 2),
      };

      const requestQueue = Object.entries(finalOutput).map(([key, value]) =>
        API.put('/api/option/', { key, value }),
      );

      const results = await Promise.all(requestQueue);

      if (results.includes(undefined)) {
        return showError(t('部分保存失败，请重试'));
      }

      for (const res of results) {
        if (!res.data.success) {
          return showError(res.data.message);
        }
      }

      showSuccess(t('保存成功'));
      props.refresh();
    } catch (error) {
      console.error('保存失败:', error);
      showError(t('保存失败，请重试'));
    } finally {
      setLoading(false);
    }
  };

  // ---- 阶梯计费辅助函数 ----
  const addTier = () => {
    setEditingTiers((prev) => [
      ...prev,
      { min_tokens: 0, max_tokens: -1, min_output_tokens: 0, max_output_tokens: 0, input: 0, output: 0 },
    ]);
  };

  const updateTier = (index, field, value) => {
    setEditingTiers((prev) =>
      prev.map((tier, i) => (i === index ? { ...tier, [field]: value ?? 0 } : tier)),
    );
  };

  const deleteTier = (index) => {
    setEditingTiers((prev) => prev.filter((_, i) => i !== index));
  };

  // 计费类型标签
  const billingModeTag = (mode) => {
    const map = {
      'per-token': { color: 'violet', label: t('按量') },
      'per-request': { color: 'teal', label: t('按次') },
      'per-second': { color: 'orange', label: t('按秒') },
      'per-tiered': { color: 'amber', label: t('阶梯') },
    };
    const cfg = map[mode] || { color: 'white', label: mode };
    return (
      <Tag color={cfg.color} shape='circle' size='small'>
        {cfg.label}
      </Tag>
    );
  };

  const columns = [
    {
      title: t('模型名称'),
      dataIndex: 'name',
      key: 'name',
      render: (text, record) => (
        <div style={{ display: 'flex', alignItems: 'center', flexWrap: 'wrap', gap: '4px' }}>
          <span>{text}</span>
          {record.hasConflict && (
            <Tag color='red' shape='circle' size='small'>
              {t('矛盾')}
            </Tag>
          )}
          {record.groupPricingGroups && record.groupPricingGroups.length > 0 && (
            <Tag color='cyan' shape='circle' size='small' style={{ cursor: 'help' }}>
              {t('分组')} ({record.groupPricingGroups.length})
            </Tag>
          )}
        </div>
      ),
    },
    {
      title: t('计费类型'),
      dataIndex: 'billingMode',
      key: 'billingMode',
      width: 80,
      render: (mode) => billingModeTag(mode),
    },
    {
      title: t('模型固定价格'),
      dataIndex: 'price',
      key: 'price',
      render: (text, record) => {
        if (record.billingMode === 'per-tiered') {
          return (
            <div>
              <span style={{ color: 'var(--semi-color-text-2)', fontSize: 12 }}>
                {(record.tiers || []).length} {t('个阶梯')}
              </span>
              {record.groupPricingGroups && record.groupPricingGroups.length > 0 && (
                <div style={{ fontSize: 11, color: 'var(--semi-color-primary)', marginTop: 2 }}>
                  {t('已配置')} {record.groupPricingGroups.length} {t('个分组')}
                </div>
              )}
            </div>
          );
        }
        return (
          <div>
            <Input
              value={text}
              placeholder={t('按量计费')}
              onChange={(value) => updateModel(record.name, 'price', value)}
            />
            {record.groupPricingGroups && record.groupPricingGroups.length > 0 && (
              <div style={{ fontSize: 11, color: 'var(--semi-color-primary)', marginTop: 4 }}>
                {t('已配置')} {record.groupPricingGroups.length} {t('个分组')}
              </div>
            )}
          </div>
        );
      },
    },
    {
      title: t('模型倍率'),
      dataIndex: 'ratio',
      key: 'ratio',
      render: (text, record) => {
        if (record.billingMode === 'per-tiered') return '-';
        return (
          <div>
            <Input
              value={text}
              placeholder={record.price !== '' ? t('模型倍率') : t('默认补全倍率')}
              disabled={record.price !== ''}
              onChange={(value) => updateModel(record.name, 'ratio', value)}
            />
            {record.groupPricingGroups && record.groupPricingGroups.length > 0 && (
              <div style={{ fontSize: 11, color: 'var(--semi-color-primary)', marginTop: 4 }}>
                {t('已配置')} {record.groupPricingGroups.length} {t('个分组')}
              </div>
            )}
          </div>
        );
      },
    },
    {
      title: t('补全倍率'),
      dataIndex: 'completionRatio',
      key: 'completionRatio',
      render: (text, record) => {
        if (record.billingMode === 'per-tiered') return '-';
        return (
          <div>
            <Input
              value={text}
              placeholder={record.price !== '' ? t('补全倍率') : t('默认补全倍率')}
              disabled={record.price !== ''}
              onChange={(value) =>
                updateModel(record.name, 'completionRatio', value)
              }
            />
            {record.groupPricingGroups && record.groupPricingGroups.length > 0 && (
              <div style={{ fontSize: 11, color: 'var(--semi-color-primary)', marginTop: 4 }}>
                {t('已配置')} {record.groupPricingGroups.length} {t('个分组')}
              </div>
            )}
          </div>
        );
      },
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
            icon={<IconLayers />}
            onClick={() => openGroupPricing(record.name)}
            title={t('分组定价')}
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

  // 阶梯配置表格列（弹窗内）
  const tierTableColumns = [
    {
      title: t('输入最小Tokens'),
      dataIndex: 'min_tokens',
      key: 'min_tokens',
      width: 120,
      render: (text, record, index) => (
        <InputNumber
          value={text}
          onChange={(value) => updateTier(index, 'min_tokens', value)}
          placeholder='0'
          style={{ width: '100%' }}
        />
      ),
    },
    {
      title: t('输入最大Tokens'),
      dataIndex: 'max_tokens',
      key: 'max_tokens',
      width: 120,
      render: (text, record, index) => (
        <InputNumber
          value={text}
          onChange={(value) => updateTier(index, 'max_tokens', value)}
          placeholder='-1'
          style={{ width: '100%' }}
        />
      ),
    },
    {
      title: t('输出最小Tokens'),
      dataIndex: 'min_output_tokens',
      key: 'min_output_tokens',
      width: 120,
      render: (text, record, index) => (
        <InputNumber
          value={text ?? 0}
          onChange={(value) => updateTier(index, 'min_output_tokens', value ?? 0)}
          placeholder='0'
          style={{ width: '100%' }}
        />
      ),
    },
    {
      title: t('输出最大Tokens'),
      dataIndex: 'max_output_tokens',
      key: 'max_output_tokens',
      width: 120,
      render: (text, record, index) => (
        <InputNumber
          value={text ?? 0}
          onChange={(value) => updateTier(index, 'max_output_tokens', value ?? 0)}
          placeholder='0'
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
      width: 70,
      render: (_, record, index) => (
        <Button
          icon={<IconDelete />}
          type='danger'
          size='small'
          onClick={() => deleteTier(index)}
        />
      ),
    },
  ];

  const updateModel = (name, field, value) => {
    if (isNaN(value)) {
      showError('请输入数字');
      return;
    }
    setModels((prev) =>
      prev.map((model) => {
        if (model.name !== name) return model;
        const updated = { ...model, [field]: value };
        updated.hasConflict =
          updated.price !== '' &&
          (updated.ratio !== '' || updated.completionRatio !== '');
        return updated;
      }),
    );
  };

  const deleteModel = (name) => {
    setModels((prev) => prev.filter((model) => model.name !== name));
  };

  const calculateRatioFromTokenPrice = (tokenPrice) => {
    return tokenPrice / 2;
  };

  // ======= 分组定价相关函数 =======
  const getAvailableGroups = () => {
    try {
      const groupRatio = JSON.parse(props.options?.GroupRatio || '{}');
      return Object.keys(groupRatio);
    } catch {
      return [];
    }
  };

  // 获取当前正在打开的模型的计费类型
  const getModelBillingMode = (modelName) => {
    const m = models.find((x) => x.name === modelName);
    return m ? m.billingMode : 'per-token';
  };

  const openGroupPricing = (modelName) => {
    // 从 GroupModelRatio 中读取该模型的现有配置
    let existingData = {};
    try {
      const gm = JSON.parse(props.options?.GroupModelRatio || '{}');
      for (const [group, modelMap] of Object.entries(gm)) {
        if (modelMap && modelMap[modelName] !== undefined) {
          existingData[group] = modelMap[modelName];
        }
      }
    } catch {}

    const groups = getAvailableGroups();
    const billingMode = getModelBillingMode(modelName);

    // 全局默认价格（用于回填 placeholder 和无分组配置时的默认值）
    const globalModel = models.find((x) => x.name === modelName);
    const globalDefaultInput = globalModel
      ? +(globalModel.ratio * 2).toFixed(6)
      : null;
    const globalDefaultOutput = globalModel && globalModel.ratio
      ? +(globalModel.ratio * (globalModel.completionRatio || 1) * 2).toFixed(6)
      : null;
    const globalDefaultPrice = globalModel ? globalModel.price : null;
    const globalDefaultTiers = globalModel ? (globalModel.tiers || []) : [];

    const makeEmpty = () => {
      if (billingMode === 'per-token') return { input: '', output: '', _globalInput: globalDefaultInput, _globalOutput: globalDefaultOutput };
      if (billingMode === 'per-request' || billingMode === 'per-second') return { price: '', _globalPrice: globalDefaultPrice };
      if (billingMode === 'per-tiered') return { tiers: [] };
      return {};
    };

    const rows = groups.map((g) => {
      const existing = existingData[g];
      if (existing === undefined || existing === null) {
        return { group: g, ...makeEmpty() };
      }
      if (billingMode === 'per-token') {
        return {
          group: g,
          input: existing.input != null ? String(existing.input) : '',
          output: existing.output != null ? String(existing.output) : '',
          _globalInput: globalDefaultInput,
          _globalOutput: globalDefaultOutput,
        };
      }
      if (billingMode === 'per-request' || billingMode === 'per-second') {
        return {
          group: g,
          price: existing.price != null ? String(existing.price) : '',
          _globalPrice: globalDefaultPrice,
        };
      }
      if (billingMode === 'per-tiered') {
        return {
          group: g,
          tiers: Array.isArray(existing.tiers) ? existing.tiers.map((t) => ({ ...t })) : [],
        };
      }
      return { group: g };
    });

    // 已配置但不在当前 groupRatio 里的分组也补充
    for (const g of Object.keys(existingData)) {
      if (!groups.includes(g)) {
        const existing = existingData[g];
        if (billingMode === 'per-token') {
          rows.push({
            group: g,
            input: existing.input != null ? String(existing.input) : '',
            output: existing.output != null ? String(existing.output) : '',
            _globalInput: globalDefaultInput,
            _globalOutput: globalDefaultOutput,
          });
        } else if (billingMode === 'per-tiered') {
          rows.push({ group: g, tiers: Array.isArray(existing.tiers) ? existing.tiers.map((t) => ({ ...t })) : [] });
        } else {
          rows.push({ group: g, price: existing.price != null ? String(existing.price) : '', _globalPrice: globalDefaultPrice });
        }
      }
    }

    setGroupPricingModel(modelName);
    setGroupPricingRows(rows);
    // 保存全局阶梯供 per-tiered 展示参考
    setGroupPricingGlobalTiers(globalDefaultTiers);
    setGroupPricingVisible(true);
  };

  const saveGroupPricing = async () => {
    setGroupPricingSaving(true);
    try {
      let current = {};
      try {
        current = JSON.parse(props.options?.GroupModelRatio || '{}');
      } catch {}

      const billingMode = getModelBillingMode(groupPricingModel);

      for (const row of groupPricingRows) {
        if (billingMode === 'per-token') {
          const inp = row.input !== '' && row.input != null ? parseFloat(row.input) : null;
          const out = row.output !== '' && row.output != null ? parseFloat(row.output) : null;
          if ((inp === null || isNaN(inp)) && (out === null || isNaN(out))) {
            // 清空
            if (current[row.group]) {
              delete current[row.group][groupPricingModel];
              if (Object.keys(current[row.group]).length === 0) delete current[row.group];
            }
          } else {
            if (!current[row.group]) current[row.group] = {};
            const entry = {};
            if (inp !== null && !isNaN(inp)) entry.input = inp;
            if (out !== null && !isNaN(out)) entry.output = out;
            current[row.group][groupPricingModel] = entry;
          }
        } else if (billingMode === 'per-tiered') {
          const tiers = Array.isArray(row.tiers) ? row.tiers.filter(
            (t) => t.min_tokens != null && t.max_tokens != null && t.input != null && t.output != null,
          ) : [];
          if (tiers.length === 0) {
            if (current[row.group]) {
              delete current[row.group][groupPricingModel];
              if (Object.keys(current[row.group]).length === 0) delete current[row.group];
            }
          } else {
            if (!current[row.group]) current[row.group] = {};
            current[row.group][groupPricingModel] = { tiers };
          }
        } else {
          const p = row.price !== '' && row.price != null ? parseFloat(row.price) : null;
          if (p === null || isNaN(p)) {
            if (current[row.group]) {
              delete current[row.group][groupPricingModel];
              if (Object.keys(current[row.group]).length === 0) delete current[row.group];
            }
          } else {
            if (!current[row.group]) current[row.group] = {};
            current[row.group][groupPricingModel] = { price: p };
          }
        }
      }

      const jsonStr = JSON.stringify(current);
      const res = await API.put('/api/option/', { key: 'GroupModelRatio', value: jsonStr });
      if (!res?.data?.success) {
        showError(res?.data?.message || t('保存失败'));
        return;
      }
      showSuccess(t('保存成功'));
      setGroupPricingVisible(false);
      props.refresh();
    } catch (e) {
      showError(t('保存失败，请重试'));
    } finally {
      setGroupPricingSaving(false);
    }
  };

  const updateGroupPricingRow = (group, field, value) => {
    setGroupPricingRows((prev) =>
      prev.map((r) => (r.group === group ? { ...r, [field]: value } : r)),
    );
  };

  // 阶梯计费：分组 tiers 操作
  const addGroupTier = (group) => {
    setGroupPricingRows((prev) =>
      prev.map((r) => {
        if (r.group !== group) return r;
        const tiers = Array.isArray(r.tiers) ? r.tiers : [];
        return {
          ...r,
          tiers: [...tiers, { min_tokens: 0, max_tokens: -1, min_output_tokens: 0, max_output_tokens: 0, input: 0, output: 0 }],
        };
      }),
    );
  };

  const updateGroupTier = (group, index, field, value) => {
    setGroupPricingRows((prev) =>
      prev.map((r) => {
        if (r.group !== group) return r;
        const tiers = (Array.isArray(r.tiers) ? r.tiers : []).map((t, i) =>
          i === index ? { ...t, [field]: value } : t,
        );
        return { ...r, tiers };
      }),
    );
  };

  const deleteGroupTier = (group, index) => {
    setGroupPricingRows((prev) =>
      prev.map((r) => {
        if (r.group !== group) return r;
        const tiers = (Array.isArray(r.tiers) ? r.tiers : []).filter((_, i) => i !== index);
        return { ...r, tiers };
      }),
    );
  };
  // ======= end 分组定价 =======

  const calculateCompletionRatioFromPrices = (
    modelTokenPrice,
    completionTokenPrice,
  ) => {
    if (!modelTokenPrice || modelTokenPrice === '0') {
      showError('模型价格不能为0');
      return '';
    }
    return completionTokenPrice / modelTokenPrice;
  };

  const handleTokenPriceChange = (value) => {
    // Use a temporary variable to hold the new state
    let newState = {
      ...(currentModel || {}),
      tokenPrice: value,
      ratio: 0,
    };

    if (!isNaN(value) && value !== '') {
      const tokenPrice = parseFloat(value);
      const ratio = calculateRatioFromTokenPrice(tokenPrice);
      newState.ratio = ratio;
    }

    // Set the state with the complete updated object
    setCurrentModel(newState);
  };

  const handleCompletionTokenPriceChange = (value) => {
    // Use a temporary variable to hold the new state
    let newState = {
      ...(currentModel || {}),
      completionTokenPrice: value,
      completionRatio: 0,
    };

    if (!isNaN(value) && value !== '' && currentModel?.tokenPrice) {
      const completionTokenPrice = parseFloat(value);
      const modelTokenPrice = parseFloat(currentModel.tokenPrice);

      if (modelTokenPrice > 0) {
        const completionRatio = calculateCompletionRatioFromPrices(
          modelTokenPrice,
          completionTokenPrice,
        );
        newState.completionRatio = completionRatio;
      }
    }

    // Set the state with the complete updated object
    setCurrentModel(newState);
  };

  const addOrUpdateModel = (values) => {
    // Check if we're editing an existing model or adding a new one
    const existingModelIndex = models.findIndex(
      (model) => model.name === values.name,
    );

    if (existingModelIndex >= 0) {
      // Update existing model
      setModels((prev) =>
        prev.map((model, index) => {
          if (index !== existingModelIndex) return model;
          const updated = {
            name: values.name,
            price: values.price || '',
            ratio: values.ratio || '',
            completionRatio: values.completionRatio || '',
            billingMode: values.billingMode || 'per-token',
            tiers: values.tiers || [],
          };
          updated.hasConflict =
            updated.price !== '' &&
            (updated.ratio !== '' || updated.completionRatio !== '');
          return updated;
        }),
      );
      setVisible(false);
      showSuccess(t('更新成功'));
    } else {
      // Add new model
      // Check if model name already exists
      if (models.some((model) => model.name === values.name)) {
        showError(t('模型名称已存在'));
        return;
      }

      setModels((prev) => {
        const newModel = {
          name: values.name,
          price: values.price || '',
          ratio: values.ratio || '',
          completionRatio: values.completionRatio || '',
          billingMode: values.billingMode || 'per-token',
          tiers: values.tiers || [],
        };
        newModel.hasConflict =
          newModel.price !== '' &&
          (newModel.ratio !== '' || newModel.completionRatio !== '');
        return [newModel, ...prev];
      });
      setVisible(false);
      showSuccess(t('添加成功'));
    }
  };

  const calculateTokenPriceFromRatio = (ratio) => {
    return ratio * 2;
  };

  const resetModalState = () => {
    setCurrentModel({ billingMode: 'per-token', price: '', ratio: '', completionRatio: '', tiers: [] });
    setPricingMode('per-token');
    setPricingSubMode('ratio');
    setEditingTiers([]);
    setIsEditMode(false);
  };

  const editModel = (record) => {
    setIsEditMode(true);
    // Determine which pricing mode to use based on the model's current configuration
    let initialPricingMode = 'per-token';
    let initialPricingSubMode = 'ratio';

    if (record.billingMode === 'per-tiered') {
      initialPricingMode = 'per-tiered';
    } else if (record.billingMode === 'per-second') {
      initialPricingMode = 'per-second';
    } else if (record.billingMode === 'per-request') {
      initialPricingMode = 'per-request';
    } else {
      initialPricingMode = 'per-token';
    }

    setPricingMode(initialPricingMode);
    setPricingSubMode(initialPricingSubMode);

    // 同步阶梯数据
    setEditingTiers(record.tiers ? [...record.tiers] : []);

    const modelCopy = { ...record };

    // If the model has ratio data and we want to populate token price fields
    if (record.ratio) {
      modelCopy.tokenPrice = calculateTokenPriceFromRatio(
        parseFloat(record.ratio),
      ).toString();

      if (record.completionRatio) {
        modelCopy.completionTokenPrice = (
          parseFloat(modelCopy.tokenPrice) * parseFloat(record.completionRatio)
        ).toString();
      }
    }

    setCurrentModel(modelCopy);
    setVisible(true);

    setTimeout(() => {
      if (formRef.current) {
        const formValues = { name: modelCopy.name };
        if (initialPricingMode === 'per-request' || initialPricingMode === 'per-second') {
          formValues.priceInput = modelCopy.price;
        } else if (initialPricingMode === 'per-token') {
          formValues.ratioInput = modelCopy.ratio;
          formValues.completionRatioInput = modelCopy.completionRatio;
          formValues.modelTokenPrice = modelCopy.tokenPrice;
          formValues.completionTokenPrice = modelCopy.completionTokenPrice;
        }
        formRef.current.setValues(formValues);
      }
    }, 0);
  };

  return (
    <>
      <Card className='!rounded-2xl shadow-sm border-0 mb-4'>
        <div className='flex items-center mb-4'>
          <Avatar size='small' color='blue' className='mr-2 shadow-md'>
            <DollarSign size={16} />
          </Avatar>
          <div>
            <Text className='text-lg font-medium'>{t('模型定价配置')}</Text>
            <div className='text-xs text-gray-600'>
              {t('管理模型的计费规则和分组定价')}
            </div>
          </div>
        </div>

        <Space wrap className='mb-4'>
          <Button
            icon={<IconPlus />}
            theme='solid'
            type='primary'
            className='!rounded-lg'
            onClick={() => {
              resetModalState();
              setVisible(true);
            }}
          >
            {t('添加模型')}
          </Button>
          <Button 
            theme='solid'
            className='!rounded-lg'
            icon={<IconSave />} 
            onClick={SubmitData} 
            loading={loading}
          >
            {t('应用更改')}
          </Button>
          <Input
            prefix={<IconSearch />}
            placeholder={t('搜索模型名称')}
            value={searchText}
            onChange={(value) => {
              setSearchText(value);
              setCurrentPage(1);
            }}
            style={{ width: 220 }}
            className='!rounded-lg'
            showClear
          />
          <Checkbox
            checked={conflictOnly}
            onChange={(e) => {
              setConflictOnly(e.target.checked);
              setCurrentPage(1);
            }}
          >
            {t('仅显示矛盾倍率')}
          </Checkbox>
        </Space>

        {conflictOnly && (
          <Banner
            type='warning'
            closeIcon={null}
            icon={<IconAlertTriangle size='large' style={{ color: 'var(--semi-color-warning)' }} />}
            description={t('当前仅显示同时配置了固定价格和倍率的矛盾模型')}
            style={{ marginBottom: 12 }}
          />
        )}

        <Table
          columns={columns}
          dataSource={pagedData}
          pagination={{
            currentPage: currentPage,
            pageSize: pageSize,
            total: filteredModels.length,
            onPageChange: (page) => setCurrentPage(page),
            showTotal: true,
            showSizeChanger: false,
          }}
        />
      </Card>

      <Modal
        title={isEditMode ? t('编辑模型') : t('添加模型')}
        visible={visible}
        width={pricingMode === 'per-tiered' ? 1100 : 520}
        onCancel={() => {
          resetModalState();
          setVisible(false);
        }}
        onOk={() => {
          if (!currentModel) return;

          // 阶梯计费：独立处理
          if (pricingMode === 'per-tiered') {
            if (editingTiers.length === 0) {
              showError(t('请至少添加一个阶梯'));
              return;
            }
            addOrUpdateModel({
              ...currentModel,
              billingMode: 'per-tiered',
              tiers: editingTiers,
              price: '',
              ratio: '',
              completionRatio: '',
            });
            return;
          }

          if (currentModel) {
            const valuesToSave = { ...currentModel };

            if (
              pricingMode === 'per-token' &&
              pricingSubMode === 'token-price' &&
              currentModel.tokenPrice
            ) {
              const tokenPrice = parseFloat(currentModel.tokenPrice);
              valuesToSave.ratio = (tokenPrice / 2).toString();

              // Calculate and set completion ratio if both token prices are available
              if (
                currentModel.completionTokenPrice &&
                currentModel.tokenPrice
              ) {
                const completionPrice = parseFloat(
                  currentModel.completionTokenPrice,
                );
                const modelPrice = parseFloat(currentModel.tokenPrice);
                if (modelPrice > 0) {
                  valuesToSave.completionRatio = (
                    completionPrice / modelPrice
                  ).toString();
                }
              }
            }

            if (pricingMode === 'per-token') {
              valuesToSave.price = '';
              valuesToSave.tiers = [];
            } else {
              valuesToSave.ratio = '';
              valuesToSave.completionRatio = '';
              valuesToSave.tiers = [];
            }

            valuesToSave.billingMode = pricingMode;
            addOrUpdateModel(valuesToSave);
          }
        }}
      >
        <Form getFormApi={(api) => (formRef.current = api)}>
          <Form.Input
            field='name'
            label={t('模型名称')}
            placeholder='strawberry'
            required
            disabled={isEditMode}
            onChange={(value) =>
              setCurrentModel((prev) => ({ ...prev, name: value }))
            }
          />

          <Form.Section text={t('定价模式')}>
            <div style={{ marginBottom: '16px' }}>
              <RadioGroup
                type='button'
                value={pricingMode}
                onChange={(e) => {
                  const newMode = e.target.value;
                  setPricingMode(newMode);

                  if (currentModel) {
                    const updatedModel = { ...currentModel, billingMode: newMode };

                    if (newMode === 'per-tiered') {
                      setEditingTiers(updatedModel.tiers || []);
                    }

                    if (formRef.current) {
                      const formValues = { name: updatedModel.name };
                      if (newMode === 'per-request' || newMode === 'per-second') {
                        formValues.priceInput = updatedModel.price || '';
                      } else if (newMode === 'per-token') {
                        formValues.ratioInput = updatedModel.ratio || '';
                        formValues.completionRatioInput = updatedModel.completionRatio || '';
                        formValues.modelTokenPrice = updatedModel.tokenPrice || '';
                        formValues.completionTokenPrice = updatedModel.completionTokenPrice || '';
                      }
                      formRef.current.setValues(formValues);
                    }

                    setCurrentModel(updatedModel);
                  }
                }}
              >
                <Radio value='per-token'>{t('按量计费')}</Radio>
                <Radio value='per-request'>{t('按次计费')}</Radio>
                <Radio value='per-second'>{t('按秒计费')}</Radio>
                <Radio value='per-tiered'>{t('阶梯计费')}</Radio>
              </RadioGroup>
            </div>
          </Form.Section>

          {pricingMode === 'per-token' && (
            <>
              <Form.Section text={t('价格设置方式')}>
                <div style={{ marginBottom: '16px' }}>
                  <RadioGroup
                    type='button'
                    value={pricingSubMode}
                    onChange={(e) => {
                      const newSubMode = e.target.value;
                      const oldSubMode = pricingSubMode;
                      setPricingSubMode(newSubMode);

                      // Handle conversion between submodes
                      if (currentModel) {
                        const updatedModel = { ...currentModel };

                        // Convert between ratio and token price
                        if (
                          oldSubMode === 'ratio' &&
                          newSubMode === 'token-price'
                        ) {
                          if (updatedModel.ratio) {
                            updatedModel.tokenPrice =
                              calculateTokenPriceFromRatio(
                                parseFloat(updatedModel.ratio),
                              ).toString();

                            if (updatedModel.completionRatio) {
                              updatedModel.completionTokenPrice = (
                                parseFloat(updatedModel.tokenPrice) *
                                parseFloat(updatedModel.completionRatio)
                              ).toString();
                            }
                          }
                        } else if (
                          oldSubMode === 'token-price' &&
                          newSubMode === 'ratio'
                        ) {
                          // Ratio values should already be calculated by the handlers
                        }

                        // Update the form values
                        if (formRef.current) {
                          const formValues = {};

                          if (newSubMode === 'ratio') {
                            formValues.ratioInput = updatedModel.ratio || '';
                            formValues.completionRatioInput =
                              updatedModel.completionRatio || '';
                          } else if (newSubMode === 'token-price') {
                            formValues.modelTokenPrice =
                              updatedModel.tokenPrice || '';
                            formValues.completionTokenPrice =
                              updatedModel.completionTokenPrice || '';
                          }

                          formRef.current.setValues(formValues);
                        }

                        setCurrentModel(updatedModel);
                      }
                    }}
                  >
                    <Radio value='ratio'>{t('按倍率设置')}</Radio>
                    <Radio value='token-price'>{t('按价格设置')}</Radio>
                  </RadioGroup>
                </div>
              </Form.Section>

              {pricingSubMode === 'ratio' && (
                <>
                  <Form.Input
                    field='ratioInput'
                    label={t('模型倍率')}
                    placeholder={t('输入模型倍率')}
                    onChange={(value) =>
                      setCurrentModel((prev) => ({
                        ...(prev || {}),
                        ratio: value,
                      }))
                    }
                    initValue={currentModel?.ratio || ''}
                  />
                  <Form.Input
                    field='completionRatioInput'
                    label={t('补全倍率')}
                    placeholder={t('输入补全倍率')}
                    onChange={(value) =>
                      setCurrentModel((prev) => ({
                        ...(prev || {}),
                        completionRatio: value,
                      }))
                    }
                    initValue={currentModel?.completionRatio || ''}
                  />
                </>
              )}

              {pricingSubMode === 'token-price' && (
                <>
                  <Form.Input
                    field='modelTokenPrice'
                    label={t('输入价格')}
                    onChange={(value) => {
                      handleTokenPriceChange(value);
                    }}
                    initValue={currentModel?.tokenPrice || ''}
                    suffix={t('$/1M tokens')}
                  />
                  <Form.Input
                    field='completionTokenPrice'
                    label={t('输出价格')}
                    onChange={(value) => {
                      handleCompletionTokenPriceChange(value);
                    }}
                    initValue={currentModel?.completionTokenPrice || ''}
                    suffix={t('$/1M tokens')}
                  />
                </>
              )}
            </>
          )}

          {(pricingMode === 'per-request' || pricingMode === 'per-second') && (
            <Form.Input
              field='priceInput'
              label={
                pricingMode === 'per-second'
                  ? t('固定价格(每秒)')
                  : t('固定价格(每次)')
              }
              placeholder={
                pricingMode === 'per-second'
                  ? t('输入每秒价格')
                  : t('输入每次价格')
              }
              onChange={(value) =>
                setCurrentModel((prev) => ({
                  ...(prev || {}),
                  price: value,
                }))
              }
              initValue={currentModel?.price || ''}
            />
          )}

          {/* ===== 阶梯计费 ===== */}
          {pricingMode === 'per-tiered' && (
            <>
              <Divider margin='12px' />
              <div
                style={{
                  marginBottom: 8,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                }}
              >
                <Text strong>{t('阶梯配置')}</Text>
                <Button icon={<IconPlus />} size='small' onClick={addTier}>
                  {t('添加阶梯')}
                </Button>
              </div>
              <Table
                columns={tierTableColumns}
                dataSource={editingTiers}
                pagination={false}
                size='small'
              />
              <div
                style={{
                  marginTop: 12,
                  padding: '10px 12px',
                  background: 'var(--semi-color-fill-0)',
                  borderRadius: 6,
                  fontSize: 12,
                  color: 'var(--semi-color-text-2)',
                }}
              >
                {t('输入/输出 最小Token=0 表示无下限；最大Token=-1（输入）或0（输出）表示无上限。每行同时匹配输入和输出范围。价格单位：美元/百万Token（$/1M tokens）。')}
              </div>
            </>
          )}
        </Form>
      </Modal>

      {/* ===== 分组定价 Modal ===== */}
      {(() => {
        const gpBillingMode = groupPricingModel ? getModelBillingMode(groupPricingModel) : 'per-token';
        const isTiered = gpBillingMode === 'per-tiered';
        const isPerToken = gpBillingMode === 'per-token';
        const isPerRequest = gpBillingMode === 'per-request' || gpBillingMode === 'per-second';

        const gpColumns = [
          {
            title: t('用户分组'),
            dataIndex: 'group',
            key: 'group',
            width: 120,
            render: (text) => <Tag color='blue' shape='circle'>{text}</Tag>,
          },
        ];

        if (isPerToken) {
          gpColumns.push(
            {
              title: t('输入价格 ($/1M tokens)'),
              dataIndex: 'input',
              key: 'input',
              render: (text, record) => (
                <InputNumber
                  value={text === '' || text == null ? null : Number(text)}
                  placeholder={record._globalInput != null ? `${t('全局')}: ${record._globalInput}` : t('留空=全局')}
                  min={0}
                  step={0.01}
                  style={{ width: '100%' }}
                  onChange={(val) =>
                    updateGroupPricingRow(record.group, 'input', val === null || val === undefined ? '' : String(val))
                  }
                />
              ),
            },
            {
              title: t('输出价格 ($/1M tokens)'),
              dataIndex: 'output',
              key: 'output',
              render: (text, record) => (
                <InputNumber
                  value={text === '' || text == null ? null : Number(text)}
                  placeholder={record._globalOutput != null ? `${t('全局')}: ${record._globalOutput}` : t('留空=全局')}
                  min={0}
                  step={0.01}
                  style={{ width: '100%' }}
                  onChange={(val) =>
                    updateGroupPricingRow(record.group, 'output', val === null || val === undefined ? '' : String(val))
                  }
                />
              ),
            },
          );
        } else if (isPerRequest) {
          gpColumns.push({
            title: gpBillingMode === 'per-second' ? t('价格（每秒）') : t('价格（每次）'),
            dataIndex: 'price',
            key: 'price',
            render: (text, record) => (
              <InputNumber
                value={text === '' || text == null ? null : Number(text)}
                placeholder={record._globalPrice != null ? `${t('全局')}: ${record._globalPrice}` : t('留空=全局')}
                min={0}
                step={0.001}
                style={{ width: '100%' }}
                onChange={(val) =>
                  updateGroupPricingRow(record.group, 'price', val === null || val === undefined ? '' : String(val))
                }
              />
            ),
          });
        }

        return (
          <Modal
            title={
              <span>
                <IconLayers style={{ marginRight: 6 }} />
                {t('分组定价')} — {groupPricingModel}
                {gpBillingMode && (
                  <span style={{ marginLeft: 8 }}>{billingModeTag(gpBillingMode)}</span>
                )}
              </span>
            }
            visible={groupPricingVisible}
            width={isTiered ? 1100 : (isPerToken ? 640 : 480)}
            onCancel={() => setGroupPricingVisible(false)}
            onOk={saveGroupPricing}
            okButtonProps={{ loading: groupPricingSaving }}
            okText={t('保存')}
            cancelText={t('取消')}
          >
            {isTiered ? (
              <>
                <div
                  style={{
                    marginBottom: 12,
                    padding: '8px 12px',
                    background: 'var(--semi-color-fill-0)',
                    borderRadius: 6,
                    fontSize: 12,
                    color: 'var(--semi-color-text-2)',
                  }}
                >
                  {t('为每个用户分组配置阶梯价格（$/1M tokens）。配置后将覆盖该分组的全局阶梯定价。清空所有阶梯行则使用全局阶梯价格。')}
                </div>
                {groupPricingRows.length === 0 ? (
                  <div style={{ textAlign: 'center', color: 'var(--semi-color-text-3)', padding: '24px 0' }}>
                    {t('暂无可用用户分组，请先在分组倍率设置中添加分组')}
                  </div>
                ) : (
                  groupPricingRows.map((row) => (
                    <div key={row.group} style={{ marginBottom: 20 }}>
                      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 6 }}>
                        <span style={{ fontWeight: 600, fontSize: 13 }}>{row.group}</span>
                        <Button
                          size='small'
                          theme='light'
                          type='primary'
                          onClick={() => addGroupTier(row.group)}
                        >
                          {t('添加阶梯')}
                        </Button>
                      </div>
                      {Array.isArray(row.tiers) && row.tiers.length > 0 ? (
                        <Table
                          size='small'
                          pagination={false}
                          dataSource={row.tiers.map((t, i) => ({ ...t, _idx: i }))}
                          rowKey='_idx'
                          columns={[
                            {
                              title: t('输入最小Tokens'),
                              dataIndex: 'min_tokens',
                              width: 110,
                              render: (val, record) => (
                                <InputNumber
                                  size='small'
                                  value={val}
                                  min={0}
                                  step={1000}
                                  style={{ width: '100%' }}
                                  onChange={(v) => updateGroupTier(row.group, record._idx, 'min_tokens', v ?? 0)}
                                />
                              ),
                            },
                            {
                              title: t('输入最大Tokens'),
                              dataIndex: 'max_tokens',
                              width: 110,
                              render: (val, record) => (
                                <InputNumber
                                  size='small'
                                  value={val}
                                  min={-1}
                                  step={1000}
                                  style={{ width: '100%' }}
                                  onChange={(v) => updateGroupTier(row.group, record._idx, 'max_tokens', v ?? -1)}
                                />
                              ),
                            },
                            {
                              title: t('输出最小Tokens'),
                              dataIndex: 'min_output_tokens',
                              width: 110,
                              render: (val, record) => (
                                <InputNumber
                                  size='small'
                                  value={val ?? 0}
                                  min={0}
                                  step={1000}
                                  style={{ width: '100%' }}
                                  onChange={(v) => updateGroupTier(row.group, record._idx, 'min_output_tokens', v ?? 0)}
                                />
                              ),
                            },
                            {
                              title: t('输出最大Tokens'),
                              dataIndex: 'max_output_tokens',
                              width: 110,
                              render: (val, record) => (
                                <InputNumber
                                  size='small'
                                  value={val ?? 0}
                                  min={0}
                                  step={1000}
                                  style={{ width: '100%' }}
                                  onChange={(v) => updateGroupTier(row.group, record._idx, 'max_output_tokens', v ?? 0)}
                                />
                              ),
                            },
                            {
                              title: t('输入价格 ($/1M)'),
                              dataIndex: 'input',
                              width: 120,
                              render: (val, record) => (
                                <InputNumber
                                  size='small'
                                  value={val}
                                  min={0}
                                  step={0.1}
                                  precision={4}
                                  style={{ width: '100%' }}
                                  onChange={(v) => updateGroupTier(row.group, record._idx, 'input', v ?? 0)}
                                />
                              ),
                            },
                            {
                              title: t('输出价格 ($/1M)'),
                              dataIndex: 'output',
                              width: 120,
                              render: (val, record) => (
                                <InputNumber
                                  size='small'
                                  value={val}
                                  min={0}
                                  step={0.1}
                                  precision={4}
                                  style={{ width: '100%' }}
                                  onChange={(v) => updateGroupTier(row.group, record._idx, 'output', v ?? 0)}
                                />
                              ),
                            },
                            {
                              title: '',
                              dataIndex: '_idx',
                              width: 50,
                              render: (idx, record) => (
                                <Button
                                  size='small'
                                  type='danger'
                                  theme='borderless'
                                  icon={<IconMinus />}
                                  onClick={() => deleteGroupTier(row.group, record._idx)}
                                />
                              ),
                            },
                          ]}
                        />
                      ) : (
                        <div>
                          <div style={{ fontSize: 12, color: 'var(--semi-color-text-3)', padding: '4px 0 6px' }}>
                            {t('无分组阶梯配置，将使用全局阶梯价格')}{groupPricingGlobalTiers.length > 0 ? t('（如下）') : ''}
                          </div>
                          {groupPricingGlobalTiers.length > 0 && (
                            <Table
                              size='small'
                              pagination={false}
                              dataSource={groupPricingGlobalTiers.map((t, i) => ({ ...t, _i: i }))}
                              rowKey='_i'
                              style={{ opacity: 0.6 }}
                              columns={[
                                { title: t('输入最小Tokens'), dataIndex: 'min_tokens', width: 110 },
                                { title: t('输入最大Tokens'), dataIndex: 'max_tokens', width: 110 },
                                { title: t('输出最小Tokens'), dataIndex: 'min_output_tokens', width: 110, render: (v) => v ?? 0 },
                                { title: t('输出最大Tokens'), dataIndex: 'max_output_tokens', width: 110, render: (v) => v ?? 0 },
                                { title: t('输入价格 ($/1M)'), dataIndex: 'input', width: 120 },
                                { title: t('输出价格 ($/1M)'), dataIndex: 'output', width: 120 },
                              ]}
                            />
                          )}
                        </div>
                      )}
                    </div>
                  ))
                )}
              </>
            ) : (
              <>
                <div
                  style={{
                    marginBottom: 12,
                    padding: '8px 12px',
                    background: 'var(--semi-color-fill-0)',
                    borderRadius: 6,
                    fontSize: 12,
                    color: 'var(--semi-color-text-2)',
                  }}
                >
                  {isPerToken
                    ? t('填写输入/输出价格（$/1M tokens）覆盖该分组的全局定价。留空则使用全局价格。此价格会再乘以该分组的分组倍率（GroupRatio）。')
                    : t('填写固定价格覆盖该分组的全局定价。留空则使用全局价格。此价格会再乘以该分组的分组倍率（GroupRatio）。')}
                </div>
                <Table
                  size='small'
                  pagination={false}
                  dataSource={groupPricingRows}
                  rowKey='group'
                  columns={gpColumns}
                />
                {groupPricingRows.length === 0 && (
                  <div style={{ textAlign: 'center', color: 'var(--semi-color-text-3)', padding: '24px 0' }}>
                    {t('暂无可用用户分组，请先在分组倍率设置中添加分组')}
                  </div>
                )}
              </>
            )}
          </Modal>
        );
      })()}
    </>
  );
}
